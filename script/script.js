// Timer simple
        let timeLeft = parseInt('{{.TimeLimit}}', 10) || 60;
        const countdownEl = document.getElementById('countdown');
        
        const timer = setInterval(() => {
            timeLeft--;
            countdownEl.textContent = timeLeft;
            
            if (timeLeft <= 0) {
                clearInterval(timer);
                document.querySelector('form').submit();
            }
        }, 1000);

        // Soumettre automatiquement si toutes les cases sont remplies
        const inputs = document.querySelectorAll('input[type="text"]');
        inputs.forEach(input => {
            input.addEventListener('input', () => {
                const allFilled = Array.from(inputs).every(inp => inp.value.trim() !== '');
                if (allFilled) {
                    setTimeout(() => {
                        document.querySelector('form').submit();
                    }, 500);
                }
            });
        });

var express = require('express')
var app = express()

app.get('/', function (req, res) {
  res.send('Hello World!')
})

var server = app.listen(process.env.PORT || 3000, function () {
  var host = server.address().address
  var port = server.address().port
  console.log('App listening at https://%s:%s', host, port)
})