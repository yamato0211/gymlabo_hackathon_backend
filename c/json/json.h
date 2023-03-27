#ifndef JSON_H_INCLUDED
#define JSON_H_INCLUDED

typedef enum JSON_TYPENAME {
	string,
	number,
	boolean, //null含む
	object,
	array,
	null, //無効なjsonノード
} Json_Typename_t;

typedef enum TOKEN_TYPE {
	TK_STR,
	TK_NUM,
	TK_BOOLEAN,
	TK_BRACKET,
	TK_COLON,
	TK_COMMA,
} Token_type_t;

struct JSON;

typedef struct JSON {
	struct JSON *next;
	Json_Typename_t type;
	char *key;
	void *data;
} Json_t;

typedef struct JSON_LIST {
	Json_Typename_t type;
	int size;
	char *list[];
} JsonList_t;

typedef struct TOKEN {
	struct TOKEN *next;
	Token_type_t type;
	int size;
	char *str;
} Token_t;

JsonList_t *mallocJsonList(int len);
Json_t *newJsonNode(Json_t *next, Json_Typename_t type, char *key, void *data, int data_size);
Json_t *newStringJsonNode(Json_t *next, char *key, char *str);
void freeJson(Json_t *json);
Json_t *searchJson(Json_t *top, char *key);
Token_t *newToken(Token_t *next, Token_type_t type, int size, char *str);
Token_t *tokenize(char *json);
Json_t *analyzeJson(char *json);
JsonList_t *parseJsonList(Token_t **tok);
char *getListString(Json_t *json, int i);
char *genJsonStr(Json_t *json);
char *genJsonNodeStr(Json_t *json);
char *genJsonListStr(JsonList_t *json);
int getJsonNodeLen(Json_t *json);

#endif
