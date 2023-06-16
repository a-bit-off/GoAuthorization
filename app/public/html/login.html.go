{{define "login"}}
<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<title>Login page</title>
</head>
<body>
<form id="loginForm" name="loginForm" action="/login" method="post" class="mt-4">
<label for="login">Логин</label><br>
<input type="login" id="login" name="login" autocomplete="on"><br>
<label for="password">Пароль</label><br>
<input type="password" id="password" name="password" autocomplete="on"><br>
<br>
<button type="submit" name="submitBtn">Войти</button>
</form>
{{if . }}
<div>
{{.Message}}
</div>
{{end}}
<br>
<a href="/signup">Зарегистрироваться</a>
</body>
</html>
{{end}}