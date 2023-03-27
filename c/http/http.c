#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <errno.h>
#include "../json/json.h"
#include "http.h"

#define HEADER "Access-Control-Allow-Origin: *\r\nAccess-Control-Allow-Headers: *\r\nAccess-Control-Allow-Credentials: true\r\nAccess-Control-Allow-Methods: GET,PUT,POST,DELETE,UPDATE,OPTIONS\r\n"

#define SIZE (5*1024)

//ファイルサイズを取得する
unsigned int getFileSize(const char *path) {
	int size, read_size;
	char read_buf[SIZE];
	FILE *f;

	f = fopen(path, "rb");
	if (f == NULL) {
		return 0;
	}

	size = 0;
	do {
		read_size = fread(read_buf, 1, SIZE, f);
		size += read_size;
	} while(read_size != 0);

	fclose(f);
	return size;
}

/* リクエストメッセージを送信する。
 * sock : 接続済みのソケット
 * request_message : リクエストメッセージを格納するバッファへのアドレス
 * buf_size : そのバッファのサイズ
 * 戻り値 : 受信したデータサイズ（バイト）
 */
int recvRequestMessage(int sock, char *request_message, unsigned int buf_size) {
	int recv_size;
	recv_size = recv(sock, request_message, buf_size, 0);
	return recv_size;
}

/* リクエストメッセージを解析する（リクエスト行のみ）
 * method : メソッドを格納するバッファへのアドレス
 * target : リクエストターゲットを格納するバッファへのアドレス
 * request_message : 解析するリクエストメッセージが格納されたバッファへのアドレス
 * 戻り値 : 成功は0、失敗は-1
 */
int parseRequestMessage(char *method, char *target, char *request_message) {
	char *line;
	char *tmp_method;
	char *tmp_target;

	//リクエストメッセージの１行目のみ取得
	line = strtok(request_message, "\n");

	//" "までの文字列を取得しmethodにコピー
	tmp_method = strtok(line, " ");
	if(tmp_method == NULL) {
		printf("get method error\n");
		return -1;
	}
	strcpy(method, tmp_method);

	//次の" "までの文字列を取得しtargetにコピー
	tmp_target = strtok(NULL, " ");
	if(tmp_target == NULL) {
		printf("get target error\n");
		return -1;
	}
	strcpy(target, tmp_target);

	return 0;
}

int getBody(char *body, char *request_message) {
	char *p = request_message;
	while(!(p[0] == '\n' && p[1] == '\n')) {
		if(p[1] == '\0') {
			strcpy(body, "");
			return -1;
		}
		p++;
	}
	strcpy(body, p+2);
	return 0;
}

/* リクエストに対する処理を行う（GETのみ）
 * body : ボディを格納するバッファへのアドレス
 * file_path : リクエストターゲットに対応するファイルへのパス
 * 戻り値 : ステータスコード
 */
int getProcessing(char **body, char *file_path) {
	FILE *f;
	int file_size;

	//ファイルサイズを取得
	file_size = getFileSize(file_path);
	if(file_size == 0) {
		printf("getFileSize error\n");
		return 404;
	}

	*body = malloc(file_size);

	//ファイルを読み込んでボディとする
	f = fopen(file_path, "r");
	fread(*body, 1, file_size, f);
	fclose(f);

	return 200;
}

/* レスポンスメッセージを作成する
 * response_message : レスポンスメッセージを格納するバッファへのアドレス
 * status : ステータスコード
 * header : ヘッダフィールドを格納したバッファへのアドレス
 * body : ボディを格納したバッファへのアドレス
 * body_size : ボディのサイズ
 * 戻り値 : レスポンスメッセージのサイズ
 */
int createResponseMessage(char *response_message, int status, char *header, char *body, unsigned int body_size) {
	unsigned int no_body_len;
	unsigned int body_len;

	response_message[0] = '\0';

	if(status == 200) {
		sprintf(response_message, "HTTP/1.1 200 OK\r\n%s\r\n", header);
		no_body_len = strlen(response_message);
		body_len = body_size;
		memcpy(&response_message[no_body_len], body, body_len);
	} else if(status == 404) {
		sprintf(response_message, "HTTP/1.1 404 Not Found\r\n%s\r\n", header);
		no_body_len = strlen(response_message);
		body_len = 0;
	} else {
		printf("Not support status(%d)\n", status);
		return -1;
	}
	return no_body_len + body_len;
}

/* レスポンスメッセージを送信する
 * sock : 接続済みのソケット
 * response_message : 送信するレスポンスメッセージ
 * message_size : 送信するメッセージのサイズ
 * 戻り値 : 送信したサイズ
 */
int sendResponseMessage(int sock, char *response_message, unsigned int message_size) {
	int send_size;
	send_size = send(sock, response_message, message_size, 0);
	return send_size;
}

void showMessage(char *message, unsigned int size) {
//*
	unsigned int i;
	printf("Show Message\n\n");
	for(i = 0; i < size; i++) {
		putchar(message[i]);
	}
	printf("\n\n");
// */
}

/* HTTPサーバの処理を行う関数
 * sock : 接続済のソケット
 * 戻り値 : 0
 */
int httpServer(int sock) {
	int request_size, response_size;
	char request_message[SIZE];
	char response_message[SIZE];
	char method[SIZE];
	char target[SIZE];
	char header_field[SIZE];
	char request_body[SIZE];
	char *body;
	int status;
	unsigned int file_size;


	request_size = recvRequestMessage(sock, request_message, SIZE);
	if(request_size == -1) {
		printf("recvRequestMessage error\n");
		return 0;
	}

	if(request_size == 0) {
		printf("connection ended\n");
		return 0;
	}

	if(parseRequestMessage(method, target, request_message) == -1) {
		printf("parseRequestMessage error\n");
		return 0;
	}

	if(strcmp(method, "POST") == 0) {
		getBody(request_body, request_message);
	} else {
		status = 404;
	}

	Json_t *json = analyzeJson(request_body);
	if(json == NULL) {
		status = 404;
	}

	response_size = createResponseMessage(response_message, status, 
			header_field, body, file_size);
	if(response_size == -1) {
		printf("createResponseMessage error\n");
		return 0;
	}

	sendResponseMessage(sock, response_message, response_size);
	return 0;
}
/*
int m(void) {
	int w_addr, c_sock;
	struct sockaddr_in a_addr;
	while(1) {
	w_addr = c_sock = 0;
	w_addr = socket(AF_INET, SOCK_STREAM, 0);
	if(w_addr == -1) {
		printf("socket error\n");
		perror(strerror(errno));
		return -1;
	}

	memset(&a_addr, 0, sizeof(struct sockaddr_in));

	a_addr.sin_family = AF_INET;
	a_addr.sin_port = htons((unsigned short)SERVER_PORT);
	a_addr.sin_addr.s_addr = inet_addr(SERVER_ADDR);

	int yes = 1;
	if(setsockopt(w_addr, SOL_SOCKET, SO_REUSEADDR, (const char *)&yes, sizeof(yes)) < 0) {
		perror("setsockopt error");
		close(w_addr);
		return -1;
	}

	if(bind(w_addr, (const struct sockaddr *)&a_addr, sizeof(a_addr)) == -1) {
		printf("bind error\n");
		perror("bind");
		close(w_addr);
		return -1;
	}

	if(listen(w_addr, 3) == -1) {
		printf("listen error\n");
		perror("listen");
		close(w_addr);
		return -1;
	}

		printf("Waiting connect...\n");
		c_sock = accept(w_addr, NULL, NULL);
		if(c_sock == -1) {
			perror("accept");
			close(w_addr);
			return -1;
		}
		printf("Connected!!\n");

		httpServer(c_sock);

		close(c_sock);
	
	close(w_addr);
	}
	return 0;
}*/

int sendNotFound(int sock) {
	char response[SIZE];
	printf("hoge fuga");
	sprintf(response, "HTTP/1.1 404 NotFound\r\n%snot found.", HEADER);
	int len = strlen(response);
	return send(sock, response, len, 0);
}

int sendSuccess(int sock) {
	char response[SIZE];
	printf("hoge fuga1");
	sprintf(response, "HTTP/1.1 200 OK\r\n%s", HEADER);
	int len = strlen(response);
	return send(sock, response, len, 0);
}

int sendSuccessEmail(int sock, char* email) {
	char response[SIZE];
	printf("hoge fuga2");
	sprintf(response, "HTTP/1.1 200 OK\r\n%s\r\n%s", HEADER, email);
	int len = strlen(response);
	return send(sock, response, len, 0);
}

