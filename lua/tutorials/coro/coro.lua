f = coroutine.create(function()
    print("Hello")
    coroutine.yield()
    print("World")
end)

print("Status before resume: ", coroutine.status(f))
coroutine.resume(f)
print("Status after first resume: ", coroutine.status(f))
coroutine.resume(f)
print("Status after second resume: ", coroutine.status(f))

f2 = coroutine.create(function()
    print("F2")
    a, b = coroutine.yield()
    print("a", a, "b", b)
end)

coroutine.resume(f2)
coroutine.resume(f2, 1, 2)
print("Status f2 after resume", coroutine.status(f2))

countdown = function(a)
    return function()
        for i = a, 0, -1 do
            coroutine.yield(i)
        end
    end
end

c = coroutine.create(countdown(10))

while coroutine.status(c) ~= "dead" do
    s, i = coroutine.resume(c)
    print("s", s, "i", i)
end
