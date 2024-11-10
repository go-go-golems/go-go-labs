http.handler("GET", "/hello", function(req, res)
    local name = req.query.name or "World"
    res:set_status(200)
    res:header("Content-Type", "text/plain")
    res.body = "Hello, " .. name .. "!"
end)
