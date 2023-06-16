{{define "signup"}}
<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<title>Sign up page</title>
</head>
<body>
<form id="loginForm" name="signUpForm" action="/signup" method="post" >
<label for="name">Имя</label><br>
<input type="text" id="name" name="name" required><br>
<label for="surname">Фамилия</label><br>
<input type="text" id="surname" name="surname" required><br>
<label for="login">Логин</label><br>
<input type="login" id="login" name="login" required><br>
<label for="password">Пароль</label><br>
<input type="password" id="password" name="password" required><br>
<label for="password2">Повторите пароль</label><br>
<input type="password" id="password2" name="password2" required><br>
<br>
<button type="submit" name="submitBtn">Зарегистрироваться</button>
</form>
{{if . }}
<div>
{{.Message}}
</div>
{{end}}
</body>
</html>
{{end}}