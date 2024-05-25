
Aims
----

* Represent multiple values as a single value
* Easily pack and unpack the values (low ceremony objects)
* Return tuples from functions without requiring allocation

Syntax
------

    var x = 1, 2, 3 // pack
    var a, b, c = x // unpack
    return a, b, c // return from function

Runtime
-------

* Four value registers and a `value_width` register
* `LOAD_n` and `STORE_n` instructions
* A `Tuple` class wrapping an array of values
* Behaviour of `STORE` when `value_width != 1` is to create a `Tuple` object
  and store that value in the register

e.g. for three values:

* `LOAD_3 r1 r2 r3`  
    Pack the three registers into a tuple value by moving `r1` into
    `value[0]`, `r2` into `value[1]` and `r3` into `value[2]` and setting
    `value_width` to `3`  
    Object allocation is deferred until a `STORE` instruction is encountered
* `STORE_3 r1 r2 r3`  
    Asserts that the value is a 3-ple, either because `value_width` is equal to
    `3` or because `value_width` is equal to `1` and the dynamic type of
    `value[0]` is `Tuple` with width of `3`, in which case the tuple fields are
    expanded into `value[0..2]` and `value_width` is set to `3`  
    Unpack the tuple into registers by moving `value[0]` into `r1`, `value[1]`
    into `r2` and `value[2]` into `r3`



