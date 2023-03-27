#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "json.h"

char *genJsonNodeStr(Json_t *json) {
	char *json_node_str = NULL;
	int key_len = strlen(json->key);
	int data_len;
	char *data_str = NULL;
	if(json->type == string || json->type == number || json->type == boolean) {
		data_len = strlen((char*)json->data);
		data_str = (char*)json->data;
		json_node_str = malloc(key_len+data_len+10);
		if(json->type == string) {
			sprintf(json_node_str, "\"%s\":\"%s\"",json->key, data_str);
		} else {
			sprintf(json_node_str, "\"%s\":%s", json->key, data_str);
		}
		return json_node_str;
	} else if(json->type == object) {
		data_str = genJsonStr((Json_t*)json->data);
		data_len = strlen(data_str);
		json_node_str = malloc(key_len+data_len+10);
		sprintf(json_node_str, "\"%s\":%s", json->key, data_str);
		free(data_str);
		return json_node_str;
	} else if(json->type == array) {
		data_str = genJsonListStr((JsonList_t*)json->data);
		data_len = strlen(data_str);
		json_node_str = malloc(key_len+data_len+10);
		sprintf(json_node_str, "\"%s\":%s", json->key, data_str);
		free(data_str);
		return json_node_str;
	}
	return "";
}

int getJsonNodeLen(Json_t *json) {
	int i = 0;
	Json_t *p = json;
	while(p != NULL) {
		i++;
		p = p->next;
	}
	return i;
}

char *genJsonListStr(JsonList_t *list) {
	char *json_list_str = NULL;
	char *element_strings[list->size];
	int list_str_len = 0;

	for(int i = 0; i < list->size; i++) {
		list_str_len += strlen(list->list[i])+1;
	}

	json_list_str = malloc(sizeof(char)*list_str_len + 10);
	strcpy(json_list_str, "[");

	for(int i = 0; i < list->size; i++) {
		strcat(json_list_str, list->list[i]);
		if(i < list->size - 1)
			strcat(json_list_str, ",");
		else
			strcat(json_list_str, "]");
	}
	return json_list_str;
}


char *genJsonStr(Json_t *json) {
	char *json_str = NULL;
	int json_len = getJsonNodeLen(json);
	int json_str_len = 0;
	char **node_strings = malloc(sizeof(char*)*json_len);
	Json_t *p = json;
	
	for(int i = 0; i < json_len; i++) {
		node_strings[i] = genJsonNodeStr(p);
		json_str_len += strlen(node_strings[i]) + 1;
		p = p->next;
	}

	json_str = malloc(sizeof(char)*json_str_len + 10);
	strcpy(json_str, "{");

	for(int i = 0; i < json_len; i++) {
		strcat(json_str, node_strings[i]);
		if(i < json_len-1) {
			strcat(json_str, ",");
		} else {
			strcat(json_str, "}");
		}
		free(node_strings[i]);
	}

	return json_str;
}
