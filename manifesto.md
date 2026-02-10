# THE AIKI MANIFESTO

**A Thesis in Code**

*This is the Way.*

### The Position

We have confused "easy" with "simple." Modern languages are easy, they guess your intent, hide their state behind syntactic sugar, and prioritize writing speed over reading clarity. Aiki is not easy. It is **simple**.

Aiki asserts that true clarity arises from the deliberate removal of choice. **In an era of Information Overload, constraints are the only antidote to chaos.** We do not stand on the shoulders of giants to hide the view; we stand there to see the horizon clearly.

### The Axioms

**1. The Arithmetic of Certainty**
Aiki rejects the IEEE 754 floating-point compromise. In this system, `1/3 * 3` is exactly `1`. There is no epsilon; there is no "close enough." We trade the raw speed of the GPU for the absolute truth of Rational arithmetic. If the math is wrong, the logic is wrong.

**2. The Rejection of Invisible Rules**
There is no operator precedence. There is no `PEMDAS` to memorize. `1 + 2 * 3` is `9`, because the parser reads left-to-right, just as the machine does.

* **The Consequence:** We reject the symbols that fool you. We have removed `==`, `!=`, and `!` from the language. These operators rely on invisible binding powers. In Aiki, you must write `equal(a, b)` or `not(equal(a, b))`. We do not rely on invisible syntax rules; we rely on visible grouping.

**3. The Universality of the List**
We reject the Cambrian explosion of data types. There are no classes, structs, or records. There are only Lists and Shapes.

* **The Proof:** The Aiki Hash Map is not a black-box primitive wrapped from the host; it is a visible construction of Aiki Lists, written in Aiki itself. We do not hide complexity; we compost it.

### The Architecture of Restraint

Complexity is not a feature; it is a tax. We reject the modern tendency to solve architectural problems with syntactic sugar.

* **Control Flow:** We provide one loop (`while`), not three. We provide one conditional (`if`), not a ladder of `switch`, `case`, or `unless`. We reject the "convenience" layers that allow two programmers to write the same logic in mutually unintelligible ways.
* **The Discipline:** By refusing to guess your intent, we force you to declare it. We offer no macros to hide code, and no inheritance to hide state. The language is finished when there is nothing left to remove.

### The Lineage

Aiki is a synthesis of the foundational disciplines that shaped our understanding of computation.

**The Spirit (Ancestry)**

* **BASIC / LOGO:** The Joy. We reclaim the immediate, visceral feedback of the 8-bit era. Aiki is a creative medium where `plot(x,y)` is as fundamental as `print(x)`. We do not just compute; we draw.
* **Algol 60:** The Rigor. We accept that for logic to be sound, the context of every variable must be static and knowable (Lexical Scope).
* **C:** The Discipline. We embrace the "small language" philosophy, the conviction that a language should be holdable in the mind of a single programmer.
* **Python:** The Workbench. We adopt the doctrine that "Explicit is better than implicit." We embrace the REPL as the physicist's workbench, valuing immediate feedback over compilation cycles.
* **Go:** The Catalyst. We acknowledge Go as the muse that proved simplicity is still a virtue in the modern era.

**The Body (Mechanics)**

* **Scheme:** We follow the path of SICP, building complexity from atoms and combinations. We utilize the homoiconic List as the universal atom of construction, but we reject the Macro.
* **Smalltalk:** We enforce the discipline of strict left-to-right evaluation, rejecting the hidden state of operator precedence.
* **ML:** We adopt the pipe operator (`|>`) and the error-value convention, favoring explicit data flow over exception jumping.
* **Forth:** We adopt the strategy of the "minimal kernel." The Aiki standard library (Prelude) is written in Aiki, proving the language's completeness by using it to build its own tools.
* **Go:** The Engine. We inhabit the Go Runtime. Aiki processes are Goroutines (M:N scheduling), and we adopt `fmt` to end all formatting wars. We accept the host's physics so we can enforce our own chemistry.

### Conclusion

Aiki is not a product. It is an epistemological stance. It proves that a system built on strict constraints, Rational math, explicit grouping, and universal data structures, remains comprehensible long after the author has moved on.

It is code that does not rot, because it hides nothing.
