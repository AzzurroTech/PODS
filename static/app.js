var token = "";

document.getElementById("auth-btn").addEventListener("click", async function() {
	var username = document.getElementById("auth-username").value;
	var password = document.getElementById("auth-password").value;
	var res = await fetch("/auth", {
		method: "POST",
		headers: { "Content-Type": "application/json" },
		body: JSON.stringify({ username: username, password: password })
	});
	var data = await res.json();
	var el = document.getElementById("auth-result");
	el.textContent = JSON.stringify(data, null, 2);
	el.classList.add("visible");
	if (data.data && data.data.token) {
		token = data.data.token;
		document.getElementById("call-btn").disabled = false;
	}
});

document.getElementById("call-btn").addEventListener("click", async function() {
	var name = document.getElementById("function-select").value;
	var payload = document.getElementById("function-payload").value;
	var res = await fetch("/functions/" + name, {
		method: "POST",
		headers: { "Authorization": token, "Content-Type": "text/plain" },
		body: payload
	});
	var data = await res.json();
	var el = document.getElementById("call-result");
	el.textContent = JSON.stringify(data, null, 2);
	el.classList.add("visible");
});
