http.handler("GET", "/calculator/add", function(req, res)
    local a = tonumber(req.query.a) or 0
    local b = tonumber(req.query.b) or 0
    res:set_status(200)
    res:header("Content-Type", "application/json")
    res.body = {result = a + b}
end)

http.handler("GET", "/calculator/subtract", function(req, res)
    local a = tonumber(req.query.a) or 0
    local b = tonumber(req.query.b) or 0
    res:set_status(200)
    res:header("Content-Type", "application/json")
    res.body = {result = a - b}
end)
