http.handler("GET", "/echo", function(req, res)
    res:set_status(200)
    res:header("Content-Type", "application/json")
    res.body = {
        method = req.method,
        headers = req.headers,
        query = req.query
    }
end)

http.handler("POST", "/echo", function(req, res)
    res:set_status(200)
    res:header("Content-Type", "application/json")
    res.body = {
        method = req.method,
        headers = req.headers,
        query = req.query,
        body = req.body
    }
end)
