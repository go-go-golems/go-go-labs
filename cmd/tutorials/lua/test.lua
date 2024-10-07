local myTable = { a = 1, b = 2 }

local mt = {
    __index = function(table, key)
        return "default value"
    end
}

setmetatable(myTable, mt)

print("myTable.a = " .. myTable.a)
print("myTable.b = " .. myTable.b)
print("myTable.c = " .. myTable.c)