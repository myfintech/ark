*utils
===

You might be wondering why these `utils` have sub folders for different types of utilities.
Why not just create a `crypto` library instead of `cryptoutils`.

Golang's standard library already has a `crypto` package.
When you import something like. `import "crypto"` the package names collide and you're forced to give your package an alias like `cryptoutils` anyway.
```go
import (
    "crypto"
    cryptoutils "my/crypto"
)
```

So, I've adopted a standard of simply defining my packages with hopefully unique names. Then their import statements look much cleaner.
```go
import (
    "crypto"
    "my/cryptoutils"
)
```

If I need to refactor a package I simply use the Jetbrains refactor features to move things around ane ensure all internal usage of those libraries are covered.

