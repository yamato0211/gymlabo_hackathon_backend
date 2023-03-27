import requests

res = requests.post('http://localhost:8080/c/signup', data={"name":"hoge","password":"pass","email":"test@mail.com","image":"hoge"})
print(res.status_code)
print(res.text)
