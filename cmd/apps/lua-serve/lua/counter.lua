local counter = 0

http.handler("GET", "/counter", function(req, res)
    res:set_status(200)
    res:header("Content-Type", "application/json")
    res.body = {count = counter}
end)

http.handler("POST", "/counter/increment", function(req, res)
    counter = counter + 1
    res:set_status(200)
    res:header("Content-Type", "application/json")
    res.body = {count = counter}
end)

http.handler("POST", "/counter/reset", function(req, res)
    counter = 0
    res:set_status(200)
    res:header("Content-Type", "application/json")
    res.body = {count = counter}
end)
