#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <errno.h>
#include <stdbool.h>
#include "json.h"

JsonList_t *mallocJsonList(int len) {
	return (JsonList_t*)malloc(sizeof(JsonList_t) + sizeof(Json_t)*len);
}

Json_t *newJsonNode(Json_t *next, Json_Typename_t type, char *key, void *data, int data_size) {
	Json_t *new_node = (Json_t*)malloc(sizeof(Json_t));
	new_node->type = type;
	new_node->key = (char*)malloc(strlen(key)+1);
	strcpy(new_node->key, key);
	new_node->data = data;
	return new_node;
}

Json_t *newStringJsonNode(Json_t *next, char *key, char *str) {
	return newJsonNode(next, string, key, str, strlen(str)+1);
}

void freeJson(Json_t *json) {
	Json_t *p = json;
	int len = getJsonNodeLen(json);
	Json_t **json_node_list = malloc(sizeof(Json_t*)*len);

	for(int i = 0; i < len; i++) {
		json_node_list[i] = p;
		if(p->type == object) {
			free(p->key);
			freeJson((Json_t*)p->data);
		} else if(p->type == array) {
			free(p->key);
			JsonList_t *q = (JsonList_t*)p->data;
			for(int j = 0; j < q->size; j++) {
				free(q->list[j]);
			}
		} else {
			free(p->key);
			free(p->data);
		}
		p = p->next;
	}
	for(int i = 0; i < len; i++) {
		free(json_node_list[i]);
	}
	return;
}

Json_t *searchJson(Json_t *top, char *key) {
	Json_t *p = top;
	while(p != NULL) {
		if(strcmp(p->key, key) == 0) {
			return p;
		}
		p = p->next;
	}
	return NULL;
}

Token_t *newToken(Token_t *next, Token_type_t type, int size, char *str) {
	Token_t *new_token = (Token_t*)malloc(sizeof(Token_t));
	new_token->next = next;
	new_token->type = type;
	new_token->size = size;
	new_token->str = (char*)malloc(strlen(str)+1);
	strcpy(new_token->str, str);
	return new_token;
}

Token_t *tokenize(char* json) {
	char *p = json;
	char *q = NULL;
	int i = 0, num = 0;
	Token_t *top = NULL;
	Token_t *tok_ptr = NULL;

	if(json[0] != '{') {
		fprintf(stderr, "invalid json.\n");
		return NULL;
	}
	top = tok_ptr = newToken(NULL, TK_BRACKET, 1, p);
	p++;
	while(*p != '\0') {
		if(*p == '{' || *p == '}' || *p == '[' || *p == ']') {
			tok_ptr->next = newToken(NULL, TK_BRACKET, 1, p);
			tok_ptr = tok_ptr->next;
			p++;
		} else if(*p == ':') {
			tok_ptr->next = newToken(NULL, TK_COLON, 1, p);
			tok_ptr = tok_ptr->next;
			p++;
		} else if(*p == ',') {
			tok_ptr->next = newToken(NULL, TK_COMMA, 1, p);
			tok_ptr = tok_ptr->next;
			p++;
		} else if(*p == '"') {
			i = 1;
			while(p[i] != '"') i++;
			i--;
			tok_ptr->next = newToken(NULL, TK_STR, i, p+1);
			tok_ptr = tok_ptr->next;
			p += i+2;
		} else if(strncmp(p, "true", 4) == 0 || strncmp(p, "null", 4) == 0) {
			tok_ptr->next = newToken(NULL, TK_BOOLEAN, 4, p);
			tok_ptr = tok_ptr->next;
			p += 4;
		} else if(strncmp(p, "false", 5) == 0) {
			tok_ptr->next = newToken(NULL, TK_BOOLEAN, 5, p);
			tok_ptr = tok_ptr->next;
			p += 5;
		} else {
			errno = 0;
			num = strtol(p, &q, 10);
			if(errno == 0 && p != q) {
				tok_ptr->next = newToken(NULL, TK_NUM, q-p, p);
				tok_ptr = tok_ptr->next;
				p = q;
			} else if(*p == ' ' || *p == '\t' || *p == '\n'){
				p++;
			} else {
				fprintf(stderr, "tokenize failed.\n");
				return NULL;
			}
		}
	}
	return top;
}

char* strFromToken(Token_t *tok) {
	static int i = 0;
	char *str = NULL;
	str = (char*)malloc((tok->size) + 1);
	strncpy(str, tok->str, tok->size);
	return str;
}

Json_Typename_t tokenTypeToJsonType(Token_type_t token_type) {
	switch(token_type) {
	case TK_STR:
		return string;
		break;
	case TK_NUM:
		return number;
		break;
	case TK_BOOLEAN:
		return boolean;
		break;
	default:
		fprintf(stderr, "invalid type.\n");
		return null;
	}
}


bool isToken(Token_t *tok, Token_type_t type, char *str) {
	if(str == NULL) {
		return tok->type==type;
	} else if(strncmp(str, tok->str, tok->size) == 0) {
		return tok->type==type;
	} else {
		return false;
	}
}

JsonList_t *parseJsonList(Token_t **tok) {
	Token_t *tok_p = *tok, *t;
	int list_len = 1;
	JsonList_t *list;
	if(!isToken(tok_p, TK_BRACKET, "[")) {
		fprintf(stderr, "list parse failed.\n");
		return NULL;
	}
	tok_p = tok_p->next;
	t = tok_p;
	
	while(!isToken(t, TK_BRACKET, "]")) {
		if(isToken(t, TK_COMMA, ",")) {
			list_len++;
		}
		t = t->next;
	}

	if(isToken(tok_p->next, TK_BRACKET, "]")) {
		list = (JsonList_t*)malloc(sizeof(JsonList_t));
		list->size = 0;
		list->type = null;
		*tok = tok_p->next;
		return list;
	}

	list = malloc(sizeof(JsonList_t) + sizeof(Json_t) * list_len);
	list->type = tokenTypeToJsonType(tok_p->type);
	if(list->type == null) {
		fprintf(stderr, "list parse failed.\n");
		return NULL;
	}
	list->size = list_len;

	for(int i = 0; i < list_len; i++) {
		list->list[i] = strFromToken(tok_p);
		if(list->type != tokenTypeToJsonType(tok_p->type)) {
			fprintf(stderr, "different type appear in list.\n");
			return NULL;
		}
		tok_p = tok_p->next;
		if(!isToken(tok_p, TK_COMMA, ",") && !isToken(tok_p, TK_BRACKET, "]")) {
			fprintf(stderr, "(%d)list parse failed.\n", __LINE__);
			return NULL;
		}
		tok_p = tok_p->next;
	}
	*tok = tok_p;
	return list;
}

Json_t *parseJson(Token_t **tok_top) {	
	Json_t *json_top = NULL, *json_p = NULL;
	Token_t *tok_p = *tok_top;
	if(!isToken(tok_p, TK_BRACKET, "{")) {
		fprintf(stderr, "(%d)parse failed.\n", __LINE__);
		return NULL;
	}
	tok_p = tok_p->next;
	while(tok_p != NULL) {
		if(tok_p->next == NULL || isToken(tok_p, TK_BRACKET, "}")) { 
			//次のトークンがないのでこのトークンは必ず終端の}になるはず。
			if(isToken(tok_p, TK_BRACKET, "}")) {
				*tok_top = tok_p;
				return json_top;
			} else {
				fprintf(stderr, "parse failed.\nmaybe you forget }\n");
				return NULL;
			}
		} else if(!isToken(tok_p->next, TK_COLON, ":")) { 
			//終端でなければ キー:要素 が来るはずなので、次のトークンは必ず:のはず
			fprintf(stderr, "(%d)parse failed.\n", __LINE__);
			return NULL;
		} else if(tok_p->next->next == NULL) {
			//キー:まであるのに要素が無いのでダメーーー！！！
			fprintf(stderr, "(%d)parse failed.\n",__LINE__);
			return NULL;
		} else if(isToken(tok_p, TK_STR, NULL)){
			if(isToken(tok_p->next->next, TK_STR, NULL)) { //type string
				if(json_top == NULL) {
					json_top = json_p = newStringJsonNode(NULL, strFromToken(tok_p),
							strFromToken(tok_p->next->next));
					tok_p = tok_p->next->next->next;
				} else {
					json_p->next = newStringJsonNode(NULL, strFromToken(tok_p),
							strFromToken(tok_p->next->next));
					json_p = json_p->next;
					tok_p = tok_p->next->next->next;
				}
			} else if(isToken(tok_p->next->next, TK_NUM, NULL)) { //type number
				if(json_top == NULL) {
					char *str = strFromToken(tok_p->next->next);
					json_top = json_p = newJsonNode(NULL, number, strFromToken(tok_p), str, strlen(str));
					tok_p = tok_p->next->next->next;
				} else {
					char *str = strFromToken(tok_p->next->next);
					json_p->next = newJsonNode(NULL, number, strFromToken(tok_p), str, strlen(str));
					json_p = json_p->next;
					tok_p = tok_p->next->next->next;
				}
			} else if(isToken(tok_p->next->next, TK_BOOLEAN, NULL)) { //type boolean
				if(json_top == NULL) {
					char *str = strFromToken(tok_p->next->next);
					json_top = json_p = newJsonNode(NULL, boolean, strFromToken(tok_p), str, strlen(str));
					tok_p = tok_p->next->next->next;
				} else {
					char *str = strFromToken(tok_p->next->next);
					json_p->next = newJsonNode(NULL, boolean, strFromToken(tok_p), str, strlen(str));
					json_p = json_p->next;
					tok_p = tok_p->next->next->next;
				}
			} else if(isToken(tok_p->next->next, TK_BRACKET, "{")) { //type object
				Token_t *t;
				t = tok_p->next->next;
				if(json_top == NULL) {
					json_top = json_p = newJsonNode(NULL, object, 
							strFromToken(tok_p), (void*)parseJson(&t), sizeof(Json_t));
					tok_p = tok_p->next;
				} else {
					json_p->next = newJsonNode(NULL, object,
							strFromToken(tok_p), (void*)parseJson(&t), sizeof(Json_t));
					json_p = json_p->next;
					tok_p = t->next;
				}
			} else if(isToken(tok_p->next->next, TK_BRACKET, "[")) {
				Token_t *t;
				t = tok_p->next->next;
				JsonList_t *list;
				list = parseJsonList(&t);
				if(list == NULL) {
					fprintf(stderr, "list parse failed.\n");
					return NULL;
				}
				if(json_top == NULL) {
					json_top = json_p = newJsonNode(NULL, array, strFromToken(tok_p), 
							list, sizeof(JsonList_t) + sizeof(char*)*list->size);
				} else {
					json_p->next = newJsonNode(NULL, array, strFromToken(tok_p), 
							list, sizeof(JsonList_t) + sizeof(char*)*list->size);
					json_p = json_p->next;
				}
				tok_p = t;
			}
		}
		if(isToken(tok_p, TK_COMMA, ",")) {
			tok_p = tok_p->next;
		}
	}
	fprintf(stderr, "(%d)parse failed.\n", __LINE__);
	return NULL;
}

Json_t *analyzeJson(char *json) {
	Token_t *t = tokenize(json);
	if(t == NULL) {
		fprintf(stderr, "analyze failed.\n");
		return NULL;
	}
	return parseJson(&t);
}

char *getListString(Json_t *json, int i) {
	if(json->type != array) {
		return NULL;
	}
	JsonList_t *list = (JsonList_t*)json->data;
	return list->list[i];
}
