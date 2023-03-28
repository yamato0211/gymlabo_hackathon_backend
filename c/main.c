#include <stdio.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <string.h>
#include <stdlib.h>
#include <errno.h>
#include <postgresql/libpq-fe.h>
#include "http/http.h"
#include "json/json.h"

#define SERVER_ADDR "172.16.238.10"
#define SERVER_PORT 8001
#define SIZE (5*1024)

int main(void){
	int w_addr, c_sock;
	struct sockaddr_in a_addr;
	int request_size, response_size;
	char request_message[SIZE];
	char response_message[SIZE];
	char method[SIZE];
	char target[SIZE];
	char header_field[SIZE];
	char request_body[SIZE];
	char body[SIZE];
	int status;
	char sql[SIZE];
	char *name;
	char *email;
	char *image;
	char *password;
	Json_t *json, *p;
	PGconn *conn;
	PGresult *res;

	conn = PQconnectdb("host=172.16.238.9 port=5432 dbname=test user=user password=password");
	if(PQstatus(conn) != CONNECTION_OK) {
		fprintf(stderr, "Connection to databese failed: %s\n", PQerrorMessage(conn));
		PQfinish(conn);
		return -1;
	}

	while(1) {
main_loop:
		w_addr = c_sock = 0;
		w_addr = socket(AF_INET, SOCK_STREAM, 0);
		if(w_addr == -1) {
			printf("socket error.\n");
			return -1;
		}

		memset(&a_addr, 0, sizeof(struct sockaddr_in));

		a_addr.sin_family = AF_INET;
		a_addr.sin_port = htons((unsigned short)SERVER_PORT);
		a_addr.sin_addr.s_addr = inet_addr(SERVER_ADDR);

		int yes = 1;
		if(setsockopt(w_addr, SOL_SOCKET, SO_REUSEADDR, (const char *)&yes, sizeof(yes)) < 0) {
			perror("setsockopt");
			close(w_addr);
			return -1;
		}

		if(bind(w_addr, (const struct sockaddr*)&a_addr, sizeof(a_addr)) == -1) {
			perror("bind");
			close(w_addr);
			return -1;
		}

		if(listen(w_addr, 3) == -1) {
			perror("listen");
			close(w_addr);
			return -1;
		}

		fprintf(stderr, "Waiting connect...\n");
		c_sock = accept(w_addr, NULL, NULL);
		if(c_sock == -1) {
			perror("accept");
			close(w_addr);
			return -1;
		}
		fprintf(stderr, "Connected!!\n");

		fprintf(stderr, "recvRequestMessage call.\n");
		request_size = recvRequestMessage(c_sock, request_message, SIZE);
		fprintf(stderr, "%s\n", request_message);
		if(request_size == -1) {
			fprintf(stderr, "recvRequestMessage error\n");
			close(w_addr);
			continue;
		}

		if(request_size == 0) {
			fprintf(stderr,"connection ended.\n");	
			close(w_addr);
			continue;
		}

		fprintf(stderr, "parseRequestMessage call.\n");
		if(parseRequestMessage(method, target, request_message) == -1) {
			fprintf(stderr,"parseRequestMessage error.\n");
			close(w_addr);
			continue;
		}

		fprintf(stderr, "check method.\n");
		if(strcmp(method, "POST") == 0) {
			fprintf(stderr, "getBody call.\n");
			fprintf(stderr, "request_message:%s\n", request_message);
			if(getBody(request_body, request_message) == -1) {
				fprintf(stderr, "getBody failed.\n");
				close(w_addr);
				continue;
			}
			fprintf(stderr, "%s\n",request_body);
		} else {
			fprintf(stderr, "method is not post.\n");
			status = 404;
		}

		if(strcmp(request_body, "") == 0) {
			fprintf(stderr,"request_body is empty.\n");
			status = 404;
		} else {
			json = analyzeJson(request_body);
			if(json == NULL) {
				status = 404;
			}
		}

		if(strcmp(target, "/c/login") != 0 && strcmp(target, "/c/signup") != 0) {
			status = 404;
		}

		if(status == 404) { //targetが不正 or Json読み込み失敗 or POST以外のメソッド
			goto NotFound;
		}

		if(strcmp(target, "/c/signup") == 0) {
			p = searchJson(json, "name");
			if(p == NULL) goto NotFound;
			name = (char*)p->data;
			p = searchJson(json, "email");
			if(p == NULL) goto NotFound;
			email = (char*)p->data;			
			p = searchJson(json, "image");
			if(p == NULL) goto NotFound;
			image = (char*)p->data;
			p = searchJson(json, "password");
			if(p == NULL) goto NotFound;
			password = (char*)p->data;
			sprintf(sql, "select * from users where email='%s';", email);
			res = PQexec(conn, sql);
			if(PQntuples(res) != 0) {
				goto NotFound;
			}
			sprintf(sql, 
					"insert into users (name,email,image,password_hash) values ('%s','%s','%s','%s');", 
					name, email, image, password);
			res = PQexec(conn, sql);
			fprintf(stderr, "%s\n", PQresStatus(PQresultStatus(res)));
			sendSuccess(c_sock);
			freeJson(json);
			close(c_sock);
			close(w_addr);
			continue;
		} else if(strcmp(target, "/c/login") == 0) {
			fprintf(stderr, "/c/login\n");
			p = searchJson(json, "email");
			if(p == NULL) goto NotFound;
			email = (char*)p->data;
			p = searchJson(json, "password");
			if(p == NULL) goto NotFound;
			password = (char*)p->data;

			sprintf(sql, "select * from users where email='%s';", email);
			res = PQexec(conn, sql);
			if(PQntuples(res) == 0) {
				fprintf(stderr, "user not found.\n");
				goto NotFound;
			}
			if(strcmp(password,PQgetvalue(res, 0, PQfnumber(res, "password_hash"))) == 0) {
				sendSuccessEmail(c_sock, email);
				close(c_sock);
				close(w_addr);
				continue;
			} else {
				fprintf(stderr, "password wrong.\n");
				goto NotFound;
			}
		}
		goto NotFound;
	}
NotFound:
	printf("%d",sendNotFound(c_sock));
	printf("NotFound\n");
	close(c_sock);
	close(w_addr);
	goto main_loop;
}
