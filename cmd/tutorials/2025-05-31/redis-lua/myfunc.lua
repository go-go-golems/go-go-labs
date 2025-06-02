#!lua name=mylib
redis.register_function('hello', function(keys, args)
   return 'Hello '..args[1]..'!'
end)