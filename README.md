# Rooby

[![Build Status](https://travis-ci.org/st0012/Rooby.svg?branch=master)](https://travis-ci.org/st0012/Rooby)

Rooby is a new object oriented language written in Go.

## Install

1. You must have Golang installed
2. You must have set $GOPATH
3. Add your $GOPATH/bin into $PATH
4. Run following command 

```
$ go get github.com/st0012/Rooby
```

## Try it!
```
$ rooby ./samples/sample-1.ro
$ rooby ./samples/sample-2.ro
$ rooby ./samples/sample-3.ro
$ rooby ./samples/sample-4.ro
```

##  Sample snippet.
```ruby
class Foo
  def set(x)
    @x = x
  end
  def get
    @x
  end
end
class Bar < Foo; end
class Baz < Foo; end

bar = Bar.new
baz = Baz.new
foo = Foo.new
bar.set(10)
baz.set(1)
foo.set(5)

puts(bar.get + baz.get + foo.get) #=> 16
```

```ruby
class User
  def initialize(name, age)
    @name = name
    @age = age
  end

  def name
    @name
  end

  def age
    @age
  end

  def say_hi(user)
    puts(@name + " says hi to " + user.name)
  end

  def self.sum_age(user1, user2)
    user1.age + user2.age
  end
end

stan = User.new("Stan", 22)
john = User.new("John", 40)
puts(User.sum_age(stan, john)) #=> 62
stan.say_hi(john) #=> Stan says hi to John
```

```ruby
class JobPosition
  def initialize(name)
    @name = name
  end

  def name
    @name
  end
    
  def self.engineer
    new("Engineer")
  end
end

job = JobPosition.engineer
puts(job.name) #=> "Engineer"
```


```ruby
puts("123".class.name) #=> String
puts(123.class.name) #=> Integer
puts(true.class.name) #=> Boolean
```

#### Build a stack using Rooby

```ruby
class Stack
  def initialize
    @data = []
  end
    
  def push(x)
    @data.push(x)
  end
    
  def pop
    @data.pop
  end
    
  def top
    @data[@data.length - 1]
  end
end

s = Stack.new
s.push(1)
s.push(2)
s.push(3)
s.push(4)
s.push(10)
puts(s.pop) #=> 10
puts(s.top) #=> 4
```

## TODO

- ~~Customize initialization method~~
- ~~Inheritance~~
- ~~Definable class methods~~
- ~~Execution command~~
- ~~Improve built in method's self implementation~~ 
- Improve syntax
    - ~~method call without self~~
    - ~~remove semicolon~~
- ~~Makefile~~
- ~~Comment support~~
- ~~Primitive type class~~
    - ~~String~~
    - ~~Integer~~
    - ~~Boolean~~
- ~~Advanced data structures (Array/Hash)~~
- for loop support
- Package management
- Allow monkey patching(?)
- Basic IO
- More documentation
    - Samples
    - Feature list
- REPL(Hard to be done)
