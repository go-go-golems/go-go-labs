local http = {}

-- Helper function to set response headers
local function set_headers(response, headers)
    for k, v in pairs(headers or {}) do
        response:header(k, v)
    end
end

-- Helper function to set cookies
local function set_cookies(response, cookies)
    for k, v in pairs(cookies or {}) do
        response:set_cookie(k, v.value, v.max_age, v.path, v.domain, v.secure, v.http_only)
    end
end

-- Register a new HTTP handler
function http.handler(method, path, func)
    local stripped_path = path:gsub("^/", "")
    local handler_name = "http_handler_" .. method:lower() .. "_" .. stripped_path:gsub("/", "_")
    print("Registering handler: " .. handler_name)
    
    _G[handler_name] = function(request)
        local req = {
            method = request.method,
            path = request.path,
            headers = request.headers,
            query = request.query,
            body = request.body,
            params = request.params,
            get_cookie = function(name) return request:get_cookie(name) end
        }

        local res = {
            status = 200,
            headers = {},
            body = "",
            cookies = {},
            set_status = function(self, status) self.status = status end,
            header = function(self, key, value) self.headers[key] = value end,
            set_cookie = function(self, name, value, max_age, path, domain, secure, http_only)
                self.cookies[name] = {value = value, max_age = max_age, path = path, domain = domain, secure = secure, http_only = http_only}
            end
        }

        func(req, res)

        return {
            status = res.status,
            headers = res.headers,
            body = res.body,
            cookies = res.cookies
        }
    end
end

-- Expose the http table globally
_G.http = http

return http
