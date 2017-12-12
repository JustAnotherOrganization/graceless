package golang

// Instead of attempting to blacklist certain packages we're just going to
// whitelist a few things for now...

var whitelist = []string{
	// Keep in Alpha!!!
	`"bytes"`,
	`"fmt"`,
	`"log"`,
	`"math"`,
	`"math/big"`,
	`"math/cmplx"`,
	`"math/rand"`,
	`"strings"`,
	`"time"`,
}