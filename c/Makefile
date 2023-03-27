TARGET = main

$(TARGET): json/genJson.c json/json.c http/http.c main.c
	gcc -lpq -o $@ $+
	rm *.o
	rm json/*.o

