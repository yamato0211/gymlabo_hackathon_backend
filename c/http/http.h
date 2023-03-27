#ifndef HTTP_H_INCLUDED
#define HTTP_H_INCLUDED

int httpServer(int);
int recvRequestMessage(int sock, char* request_message, unsigned int buf_size);
int parseRequestMessage(char* method, char* target, char* request_message);
int getProcessing(char** body, char* file_path);
int createResponseMessage(char* response_message, int status, 
		char* header, char* body, unsigned int body_size);
int sendResponseMessage(int sock, char* response_message, unsigned int message_size);
unsigned int getFileSize(const char* path);
void sendNofFound(int sock);

#endif
