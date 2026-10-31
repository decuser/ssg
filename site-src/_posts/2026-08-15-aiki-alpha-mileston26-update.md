---

title: Aiki: Alpha Milestone 26 Update
tags: aiki programming-languages ai self-hosting
---

I push Aiki frequently, but the last couple of days have been unusually intense.

I return to my day job next week. I am a computer information systems professor at Tarleton State University, so I have been taking advantage of the remaining summer window to push several related pieces of the project to closure. The result is a surprisingly large delta from the first public alpha, even though the visible language has not exploded in size.

For those following Aiki development, this is a longer and more technical update on what changed, what the work exposed, and where I think the project is going.

I have tagged the current state as [`v0.4.0-alpha-26`](https://github.com/decuser/aiki/tree/v0.4.0-alpha-26).

I'm using tags to mark where it makes sense to stop and inspect the system as a whole, and to think about what has changed. I have not created a separate GitHub Release for this one; the tag is the milestone.

<!--more-->

## Where the First Alpha Started

The first public alpha already contained most of what someone would recognize as the Aiki language.

It had exact rational arithmetic, left-to-right binary evaluation, first-class functions, recursion and iteration, pipelines, shaped lists, modules, recoverable errors, graphics, turtle facilities, isolated message-passing concurrency, a standard library, semantic profiling, executable documentation, behavioral golds, grammar coverage, structural checks, and a growing invariant framework.

That matters because most of the work described below is not a story about accumulating language features.
````markdown
---

title: Aiki: Alpha Milestone 26 Update
tags: aiki programming-languages ai self-hosting
---

I have been busy with Aiki lately, and the last couple of days have been especially active. I return to my day job next week—I am a Computer Information Systems professor at Tarleton State University—so I have been using the remaining summer window to close several related pieces of work.

The current state is tagged [`v0.4.0-alpha-26`](https://github.com/decuser/aiki/tree/v0.4.0-alpha-26). I do not plan to create a GitHub Release for every such point. `master` moves as development continues; tags mark states worth stopping on.

The first public alpha already included most of the language proper: exact rational arithmetic, left-to-right evaluation, first-class functions, recursion and iteration, pipelines, shaped data, modules, recoverable errors, isolated concurrency, graphics, a standard library, semantic profiling, executable documentation, behavioral golds, grammar coverage, and a substantial invariant framework. Most of the work since then has not enlarged that surface very much. It has made the relationships underneath it considerably stricter.

<!--more-->

## Authority and the Grammar

The first substantial post-alpha work was on newline handling. `grammar.ebnfx` described the productions, but part of the actual surface rule still lived in parser code: whether a physical newline terminated a statement depended on a private completion set and delimiter-suppression logic. That policy now lives in the grammar. The parser consumes it, help exposes it, and structural analysis derives its consequences.

This exposed several useful facts about the existing language. Some tokens that initially looked like problematic continuations—`(`, `[`, and `-`—are actually resolved as beginnings of new expressions after termination. Other continuations are unambiguously blocked by the current rule. A function literal ending in `}` exposes another edge because `}` can end an expression without itself being part of the newline-completion set. I did not change those language choices merely to make the rule more symmetrical. They are now visible design questions rather than behavior hidden in parser code.

The same work tightened several grammar-sensitive relationships. Evaluator coverage is checked against the syntax the grammar can actually produce. Formatter coverage is explicit across the grammar rather than falling through silent recursion. Linter knowledge about syntax nodes is checked. Binary-operator membership comes from the grammar instead of a separate evaluator list. Newline help and related diagnostics derive from the declared policy.

A second pass then found that several consumers were independently walking the grammar to derive the same structural facts. Those derivations are now centralized in a cached grammar analysis. Production names, token references, AST-producing node types, terminal alternatives, and newline analysis are derived once and consumed where needed.

The working rule is straightforward: a fact should have one authority. Consumers may have their own policy and representation, but they should not independently reconstruct the same fact.

## Tooling and Failure Semantics

The newline work introduced deliberately unparsable smoke specimens, which exposed some weaknesses in the surrounding tools. Negative parser specimens now declare themselves explicitly:

```text
# @negative parse
````

The declaration is restricted to the test fixtures where it is meaningful. It cannot be used in ordinary source to exempt malformed code from formatting or linting.

Following that through uncovered several older problems. Recursive formatting could encounter malformed source, report it, and still fail to propagate the error correctly. Lint's formatting preflight could stop after the first malformed file. Several commands returned status codes internally that the top-level executable discarded, allowing a command to print failure and still exit successfully.

Those paths now propagate their status to the shell.

The fuller lint traversal also exposed a separate module-resolution drift. Runtime resolved public package names through the module registry, while lint maintained a filesystem approximation of that rule. Valid code using `use("list")` was enough to expose the disagreement once lint began reaching source it had previously skipped. Lint now uses the same registry model for public package resolution while retaining normal relative-path handling for path imports.

I followed the grammar project with a more general audit and began recording findings with stable identifiers and explicit dispositions. Some were fixed, some accepted, some deferred. Newline delimiter suppression, for example, now tracks expected closing delimiters rather than an aggregate depth that an unmatched closer could corrupt. The point of the ledger is not to turn every observation into work; it is to keep credible findings from disappearing when a project ends.

## Distribution and Location Independence

Aiki had also accumulated an implicit assumption that the source tree was the installation.

That worked because the development executable normally lived beside `lib/`, but it meant runtime correctness was partly dependent on where the process happened to be started. The running executable now identifies its distribution. Shipped modules are found relative to that executable, user modules have an explicit home, and named package discovery does not recursively scan arbitrary trees beneath the current working directory.

The normal installation model is therefore just:

```text
unpack Aiki
add the directory to PATH
run aiki
```

`make dist` constructs the user distribution. `make distcheck` goes further: it unpacks that distribution under a temporary prefix, runs from an unrelated directory, plants misleading package trees there, and verifies that the installed Aiki still resolves and runs its own modules correctly.

There is also a separate `make baseline` target for development snapshots. A baseline retains the repository, including `.git`, and is intended as a complete restartable development state. It is not the user distribution.

This distinction later became important again during self-hosting. Several things that worked from the repository root turned out to depend on being there.

## An Independent Aiki Front End

The largest body of post-alpha work began with a question about self-description: how much of Aiki could be implemented independently in Aiki itself?

There is now an Aiki-written lexer, newline normalizer, and recursive-descent parser under `selfhost/`. They do not call the Go lexer or parser.

Independence does require duplication of some facts. A lexer has to know the keywords and operators it recognizes. A newline normalizer has to know the completion and suppression policy. Those duplicated facts are checked against values derived from `grammar.ebnfx`; the algorithms themselves remain independent.

The lexical and newline implementations are compared with the Go implementation against reviewed conformance projections. For parsing, I reused the existing grammar-shaped parse golds rather than inventing another expected-tree format. The result is a useful three-way relationship among grammar and reviewed evidence, the Go implementation, and the Aiki implementation.

This is different from simply adding more tests around the same implementation. The second front end has its own code and its own opportunities to be wrong.

## Language Services

The editor work grew from the same concern about authority.

A parser, formatter, linter, editor extension, and LSP server can easily become several incomplete copies of the language. Rather than make the LSP server another owner of Aiki semantics, I moved reusable language knowledge into an editor-independent service layer.

That layer now provides document analysis, structured diagnostics, symbols, definition lookup, canonical formatting, completion, and hover information. `aiki lsp` is an adapter over that service rather than the service itself.

This has been exercised through three rather different editors. Xed uses its normal GtkSourceView mechanism for lexical presentation and a thin plugin for live diagnostics. VS Code uses a conventional thin language client and has been tested for diagnostics, completion, hover, Go to Definition, and Format Document. For nvi, Aiki generates ordinary tags:

```text
aiki tags -o tags source.ai
```

and nvi uses them in the usual way.

The live editor tests were useful. Xed exposed Save-As document identity and the presentation of zero-width EOF diagnostics. VS Code exposed an invalid JSON-RPC null response and an assumption that definition requests would be positioned at the first byte of an identifier. Desktop-launched editors also demonstrated that their process environment cannot be inferred from an interactive shell's `PATH`.

Those were adapter problems, not invitations to move Aiki semantics into the adapters.

## From Parser to Interpreter

The independent front end then grew into an interpreter.

Aiki now has an Aiki-written runtime environment and evaluator supporting lexical lookup, shadowing, assignment into enclosing scopes, closures, recursion, rest parameters, lists, shapes, indexing, field access, statements, control flow, pattern matching, pipelines, and returns.

This work exposed two gaps in the language surface.

The parser could read arbitrary symbol literals, but ordinary Aiki had no operation for constructing an arbitrary native symbol from a string. A self-hosted interpreter therefore could not faithfully turn a source lexeme into the corresponding value without either maintaining a finite private symbol table or introducing its own substitute representation. Neither was acceptable.

The result was:

```aiki
to_symbol("foo")
```

which produces `:foo`.

Shaped values exposed a related boundary. The substrate already knew how to construct them, but ordinary Aiki did not. That became:

```aiki
shaped(:point, [1, 2])
```

These are small additions to the language, but they came from a useful criterion: an implementation of Aiki written in Aiki should not need a private representation for values that ordinary Aiki source can express.

## Modules and the HAL Boundary

Modules required a similar decision.

It would have been easy for the self-hosted evaluator to delegate `import()` back to the host interpreter, but that would also delegate source-module semantics to the implementation being checked. The self-hosted path therefore performs Aiki-source module loading itself:

```text
resolve
read
lex
normalize
parse
evaluate
collect exports
cache
```

Platform facilities remain host capabilities. A privileged bootstrap captures the HAL functions needed by blessed library modules and installs them into the interpreted library environment without exposing those raw bindings to ordinary callers.

The distinction is intentional. Aiki source semantics can be implemented in Aiki. File access, clocks, graphics, native regex, and other platform effects still have to come from the substrate.

## Behavioral Conformance

Running existing behavior specimens through the independent evaluator found a semantic mistake that the more focused evaluator tests had missed.

Aiki recoverable errors such as:

```text
[@error, :math, "division by zero"]
```

are ordinary values. They can be bound, returned, inspected, and matched.

The self-hosted evaluator initially used `is_error()` to recognize its own internal failures, thereby conflating a recoverable Aiki value with an evaluator halt. The interpreter now uses a separate private fault representation for its own control path, leaving `[@error, ...]` values recoverable.

The behavior-conformance work now covers representative arithmetic, closures and recursion, matching, pipelines, relative imports, strings, file effects, pure and HAL-backed modules, bytes, hashes, regex behavior, and other existing language specimens. Concurrency, graphics, debugger-only cases, and interactive I/O remain outside the present self-host proof rather than being silently counted as covered.

## Recursive Self-Interpretation

The eventual target was to run an Aiki-written interpreter through itself.

The first attempts appeared to spend most of their time parsing the inner parser, which suggested that recursive-descent parsing under a tree-walking interpreter was the main problem. Semantic profiling showed otherwise. A large part of the cost was in the independent lexer's character handling: repeated string indexing caused repeated rune materialization over large source strings.

`string.chars` now has a linear substrate realization, and the self-host lexer snapshots the source once and avoids some repeated table scans and position work. That reduced the measured workload enough to expose the next problems.

Those were module problems rather than parser problems. Path fallback in the independent loader did not quite match the host loader. Different paths involving `.` and `..` could name the same physical module but become separate cache entries. Running recursively from outside the repository then exposed a further assumption: the self-hosted loader needed access to the same ordered distribution module roots as the host loader.

Those were corrected without handing the inner parser or evaluator back to Go.

The resulting path now works:

```text
Go-hosted Aiki
    -> Aiki-written interpreter
        -> self-host-loaded Aiki interpreter
            -> 1 + 2 * 3
                -> 9
```

The result is `9` because ordinary Aiki binary expressions are evaluated left to right: `(1 + 2) * 3`. The same visible semantic rule survives at the third level.

Go remains the production runtime and bootstrap substrate. The self-host implementation is a second implementation and conformance boundary, not an attempt to pretend that the underlying machine or host runtime has disappeared.

## Profiling the Profiler

Semantic profiling was already present in the first alpha. During the self-hosting work it became useful enough that the measurements themselves deserved scrutiny.

The recursive profiles are large. Rather than accept them because they looked plausible, I built a calibration experiment beginning with work whose semantic counts can be read directly from a few lines of Aiki.

The first programs put:

```aiki
1 + 1
```

inside one, two, three, and four nested loops of ten. The leaf calculation therefore executes 10, 100, 1,000, and 10,000 times. The total completed loop-body iterations are 10, 110, 1,110, and 11,110. The profiler reports those counts exactly.

Arithmetic is equally direct. Each leaf contributes one addition and each loop-body iteration increments one loop variable, so the expected arithmetic counts are 20, 210, 2,110, and 21,110. Those counts are exact as well. The observed comparison sequence—11, 121, 1,221, and 12,221—is also exactly what follows from testing each loop condition once more than its body executes.

A simpler expression:

```aiki
1 + 2 + 3 + 4
```

contains three arithmetic operations; the profiler reports three.

A conventional loop:

```aiki
while i < N {
    x = x + 1
    i = i + 1
}
```

has the directly predictable relationships:

```text
arithmetic = 2N
comparison = N + 1
iteration  = N
```

For `N` equal to 10, 100, and 1,000, the observed counts match all three relationships exactly.

The experiment then moves into self-interpretation. With one Aiki-written interpreter loaded, one additional evaluation contributes:

```text
arithmetic       958
comparison      1,082
call            3,255
iteration         832
index             935
```

Two evaluations add exactly twice that work, and four add exactly four times that work.

At the next interpreter level the fixed baseline is already large:

```text
arithmetic       689,681
comparison     1,346,671
call           5,021,109
iteration        498,039
index            793,179
```

One additional third-level evaluation then adds:

```text
arithmetic      +353,930
comparison      +388,932
call          +2,342,216
iteration       +304,091
index           +397,156
```

and the next identical evaluation adds exactly the same vector.

I ran the complete experiment twice. Every semantic counter was reproduced exactly across the two runs, including the native, one-level, and two-level self-host cases. The Go realization figures—elapsed time, cumulative allocation traffic, malloc count, and garbage-collection cycles—varied modestly, as expected.

The distinction is important. The semantic counters describe Aiki-level work. The Go measurements describe a particular realization of that work by the runtime and machine.

The calibration therefore connects programs countable by inspection with recursive interpretation involving millions of semantic events. Large recursive profiler counts are not being accepted merely because they have plausible scale; they are produced by the same instrumentation that is exact on small source and additive under repeated interpreted work.

## Experiments Are Not Tests

That calibration became Experiment 001 and, in the process, exposed the need for another kind of project artifact.

Tests and experiments answer different questions. A test normally states a relationship the implementation is required to preserve. An experiment records a procedure, an observation, and an interpretation without turning the observation into a correctness gold.

Aiki experiments therefore have a simple structure:

```text
001-profiler-calibration/
    README.md

    experiment/
        PROCEDURE.md
        run.sh
        materials...

    results/
        observations...

    analyses/
        interpretations...
```

The root orients the reader. `experiment/` records the procedure and its materials. `results/` contains what happened. `analyses/` contains what I think the observations mean.

New work can be scaffolded outside the source tree with:

```text
aiki experiment new "name"
```

The running distribution supplies the sequence number from its curated `experiments/` collection, while the experiment itself is created in the caller's current directory. Finished work can then be promoted into the repository.

The distinction among procedure, observation, and interpretation is useful enough that I expect experiments to become a regular part of Aiki development.

## The Commitments

None of this changes the original commitments of the project.

I still want Aiki to have a small surface and to make computational relationships visible enough to apprehend rather than merely execute. I still prefer direct composition over hidden machinery. The implementation still does not define the language merely by existing.

The extensive use of generative AI in the implementation makes this more important rather than less. The response is not to obscure that AI produced source code. It is to make the resulting source increasingly answerable to authorities and evidence outside itself.

The grammar owns syntax facts. Reviewed behavior constrains semantics. Executable documentation constrains examples. The language surface constrains implementation choices. Independent implementations can be compared. Editor clients exercise the service boundary. Relocated distributions test assumptions about location. Experiments retain their observations beside the procedure that produced them.

The direction is toward fewer places where an implementation assumption can quietly become a language fact.

## Platforms

Linux remains Aiki's development home. That is where it is built, changed, and exercised continuously.

Intel macOS is periodically put through the more rigorous build and validation path rather than being assumed portable merely because it compiles. Windows is also run periodically and remains part of the portability work, although neither platform receives the continuous development attention that Linux does.

Aiki is cross-platform alpha software, developed primarily on Linux.

## Pending Work

The authority work made several unresolved language questions easier to see. Newline behavior still has intentionally deferred edges, including the treatment of some leading continuations and the special case around function-literal endings. Those are now explicit language-design questions rather than accidental parser rules.

The language-service architecture is also deliberately incomplete. Incremental parsing, workspace-wide indexing, semantic tokens, rename/refactoring, and similar facilities are possible extensions, but they are not requirements merely because LSP has names for them. I would rather add them when actual use shows where the simpler model is inadequate.

Spawned abnormal termination needs further thought as well. Aiki's concurrency model makes isolation and message passing explicit; failure of spawned computations should have equally explicit observable semantics rather than becoming a leak from the Go substrate.

The broader project record now distinguishes proposals, audit findings, design decisions, experiments, bugs, and implementation sessions. I expect to continue using those distinctions rather than turning every open question into an implementation task.

## Future Direction

The architectural direction is to continue tightening responsibilities and authorities: make each fact, capability, and policy belong somewhere definite; derive rather than duplicate where possible; and make important relationships executable.

The next capability work is likely to concentrate on the systems substrate.

### Process Execution

I expect to start with process execution. A useful interface needs to launch a program with arguments, standard input, environment, and working directory, and return standard output, standard error, and exit status.

The host operation is narrow and belongs below the HAL boundary. Process orchestration belongs in Aiki. This is a small addition with a large practical effect because it allows Aiki programs to compose ordinary system tools.

### Networking

TCP requires a similarly small host surface: listen, accept, connect, read, write, and close.

Aiki already has isolated `spawn`, channels, and message passing. A concurrent server should be constructed by composing those facilities with networking rather than by adding a separate server model.

### JSON

JSON should be largely an Aiki library. The self-host lexer has already demonstrated character-level scanning, and Aiki values map reasonably well onto JSON values. A parser and emitter written in Aiki would be another useful test of where the HAL boundary belongs.

### Operating-System Surface

A systems environment also needs ordinary process facts: arguments, environment variables, current directory, directory change, exit status, process ID, hostname, and related information.

I expect these to be a collection of small HAL capabilities behind an Aiki `os` module rather than additions scattered through the prelude.

### Signals

Signal handling has to fit the concurrency model. My present direction is for signals such as SIGINT and SIGTERM to arrive as channel events, allowing ordinary message passing and `select` to coordinate them instead of introducing an unrelated asynchronous control mechanism.

### Timers

Wall-clock time, elapsed time, and sleeping are straightforward host facilities. Timed coordination should compose with concurrency; `time.after` naturally fits the existing `select` design.

### Cryptographic Hashing

SHA-256 and HMAC are another relatively clean boundary and become useful once programs begin dealing with integrity or authentication.

Across these additions the question remains the same: what is the smallest thing the host must provide, and what can Aiki construct from it?

## Aiki at `v0.4.0-alpha-26`

Much of the work since the first alpha is architectural rather than immediately visible at the prompt. The following is a compact inventory of the tagged system.

### Language

* exact rational arithmetic;
* left-to-right ordinary binary evaluation;
* numbers, booleans, runes, strings, symbols, lists, and first-class functions;
* functions, closures, recursion, rest parameters, iteration, and pipelines;
* pattern matching;
* shaped list data and named fields;
* recoverable error values;
* modules and explicit imports;
* isolated spawn-based concurrency and channels;
* graphics, canvas, and turtle facilities.

### Language Authority and Description

* declarative grammar owning the syntax surface;
* grammar-owned newline termination and suppression policy;
* centralized cached structural grammar analysis;
* grammar/evaluator coupling;
* grammar/formatter coupling;
* checked linter syntax knowledge;
* grammar-derived binary-operator membership;
* grammar-backed syntax help;
* reviewed structural parse projections;
* executable documentation.

### Implementations

* production Go-hosted interpreter;
* independent Aiki-written lexer;
* independent Aiki-written newline normalizer;
* independent Aiki-written recursive-descent parser;
* Aiki-written runtime environment;
* Aiki-written evaluator;
* self-hosted Aiki-source module resolution, loading, exports, and caching;
* scoped bootstrap bridge for HAL capabilities;
* recursive self-interpretation through a third-level Aiki program.

### Language Services and Editors

* editor-independent document analysis;
* structured diagnostics;
* symbol discovery and definition lookup;
* canonical formatting;
* completion;
* hover from source definitions and authored help;
* language-service observation and instrumentation;
* `aiki lsp`;
* Xed syntax and live diagnostics;
* VS Code diagnostics, completion, hover, definition, and formatting;
* classic `aiki tags` generation for nvi and other tags consumers.

### Modules and Platform Boundary

* standard Aiki module library;
* pure Aiki and HAL-backed modules;
* explicit distribution module roots;
* user library support;
* explicit relative imports;
* relocatable distribution lookup;
* dynamic symbol construction with `to_symbol`;
* general shaped-value construction with `shaped`.

### Validation and Conformance

* Go tests;
* Aiki-native tests;
* behavioral smokes;
* reviewed gold files;
* executable documentation;
* grammar-production coverage;
* structural engine golds;
* formatter parse-preservation checks;
* lint;
* tree integrity checking;
* fuzz tests;
* property tests;
* explicit negative-test declarations;
* cross-implementation lexer, newline, parser, and behavioral conformance;
* recursive self-interpretation invariants.

### Profiling and Experiments

* deterministic Aiki-level semantic counters;
* source attribution;
* Go runtime-realization measurements;
* recursive self-host profiling;
* empirical profiler calibration;
* numbered reproducible experiment homes;
* separate procedure, raw results, and analyses;
* Experiment 001, profiler calibration.

### Distribution and Development

* `make validate`;
* `make dist`;
* `make distcheck`;
* restartable repository baselines;
* relocatable user distributions;
* structural checking of distributed artifacts;
* proposals;
* audit findings;
* recorded design decisions;
* in-repository AI engineering provenance and restart records.

## Where This Leaves Aiki

The first public alpha established that Aiki had become a usable small language. Milestone 26 is less about adding another layer of surface features than about making more of the system accountable to explicit relationships.

The grammar now owns more of the syntax it describes. Several formerly duplicated facts have identifiable authorities. The editor integrations consume common language services. A second implementation now covers the front end, evaluator, and module path. Aiki can run that implementation recursively. Distribution and self-hosting have both been exercised away from the source-tree environment that originally sheltered them. The profiler has been calibrated from directly countable programs through recursive interpretation.

There is plenty left to do, but this is a useful state to mark.

Development continues on `master`; `v0.4.0-alpha-26` is the milestone.

### Links

* [Aiki on GitHub](https://github.com/decuser/aiki)
* [Aiki `v0.4.0-alpha-26`](https://github.com/decuser/aiki/tree/v0.4.0-alpha-26)
* [First Public Alpha](/posts/aiki-alpha-release/)

*post added 2026-08-15 19:49:00 -0500*

```
```

The first alpha demonstrated that Aiki had become a usable small language. The work since then has been increasingly concerned with a harder problem: making the language, its implementation, its tools, and even its measurements answerable to explicit and independently testable relationships.

The principle was already present in the alpha.

The implementation is not the specification.

I conceived, architected, designed, specified, and directed Aiki. Generative AI has been used extensively as an implementation instrument. AI systems have produced substantial amounts of source code, tests, tooling, and engineering material under that direction.

Because of that relationship, I have been particularly unwilling to accept "the implementation does this" as a definition of what the language means.

The language has to live somewhere else: in its visible behavior, grammar, documented semantics, examples, tests, gold files, invariants, and other externally imposed criteria. The generated implementation is replaceable. The language is not whatever the current Go happens to do.

Much of the work since the alpha has tightened that commitment.

## Making Authority Explicit

One of the first post-alpha projects concerned something apparently mundane: newlines.

Aiki's grammar described the language, but part of the rule determining whether a physical newline terminated a statement still lived privately in parser code. That meant the grammar was not quite the whole account of the syntax a programmer actually encountered.

The immediate problem could have been fixed narrowly.

Instead, the grammar became the authority for the newline token, the token classes and lexemes after which a newline may terminate a statement, and the delimiter pairs inside which newline termination is suppressed.

The parser now consumes that declaration.

That change made it possible to derive structural consequences from the grammar rather than merely reason about them informally: which forms can end expressions, which can continue them, where continuations are ambiguous, which continuations the present newline rule blocks, and where the grammar exposes unresolved design choices.

The same project tightened several other relationships.

Evaluator coverage is checked against the actual grammar-produced syntax surface. Formatter coverage is explicit. Linter knowledge about syntax nodes is checked. Binary-operator membership comes from the grammar rather than from a second private list in evaluator helpers. Newline help comes from grammar metadata. Diagnostics can explain a newline decision using the rule that actually caused it.

This does not mean that the grammar owns everything.

Syntax owns syntax. The evaluator still owns meaning. The formatter still owns formatting policy. The linter still owns the analyses appropriate to linting.

The distinction I have settled on is narrower:

> A fact should have one authority. Consumers may interpret that fact for their own purposes, but they should not independently reinvent it.

That led directly to another change.

Several components were walking the grammar independently to derive the same structural information. Those derivations are now centralized in one cached grammar analysis. Production names, token references, AST-producing node types, terminal alternatives, and newline analysis are derived once and consumed where needed.

The rule is simple:

> Parse once. Derive shared structural facts once. Consume everywhere.

I expect to keep tightening that noose.

## Making Failure Tell the Truth

The grammar work also introduced deliberately invalid parser specimens.

That exposed weaknesses in the tools themselves.

An intentionally unparsable smoke test and an accidentally malformed source file are not the same thing, so negative specimens now declare their intent explicitly:

```text
# @negative parse
````

That declaration is itself constrained. It is valid only in the test context where such a specimen makes sense. It cannot be dropped into arbitrary source as a magic escape from formatter or linter scrutiny.

Following that thread uncovered more consequential problems.

Recursive formatting could encounter malformed ordinary source, print the failure, and fail to propagate it correctly. Lint's formatting preflight could stop at the first malformed file. Several subcommands could return a failure status internally while the top-level executable discarded that status and exited successfully.

Those paths now tell the truth at the process boundary.

If `fmt`, `lint`, `smoke`, `enginesmoke`, `debug`, `profile`, `treecheck`, or another status-returning command says it failed, the shell gets a failure.

The fuller traversal also exposed an older linter drift. Runtime resolved public modules through the module registry; lint had a separate filesystem approximation of the same rule. Once lint actually reached previously hidden valid source, the two disagreed.

The private resolver was removed in favor of the same module-registry model used by runtime.

Again, the interesting part was not the individual bug.

The recurring question was: **who owns this fact?**

## Auditing the Exceptional Paths

After the grammar work, I did not immediately add more capability. I audited what had just been built.

That produced a findings ledger with stable identifiers and explicit dispositions: resolved, accepted, deferred.

Not every observation becomes a bug. Not every review comment becomes work. But neither should a credible finding quietly disappear because the immediate project has moved on.

That audit hardened several exceptional paths.

Newline suppression now tracks expected closing delimiters rather than one aggregate depth that an unmatched closer could corrupt. If grammar-derived newline analysis is unavailable, that degradation can be observed by diagnostic-aware clients rather than vanishing silently. Negative test declarations were narrowed to their legitimate scope. Repeated lint candidate classification was consolidated.

Some findings were deliberately left as design questions.

That distinction: fix, accept, or defer, is becoming as important to me as finding the issue in the first place.

## Separating Aiki from Its Source Tree

Another assumption had been hiding in plain sight.

Aiki naturally worked from its source tree. The executable and `lib/` directory were together, so module discovery appeared straightforward.

But the development layout had become an implicit installation model.

That is not a real distribution boundary.

The running executable now identifies its distribution. Shipped libraries are resolved relative to it, not by recursively wandering whatever directory happens to be current.

The normal installation model is consequently very small:

```text
unpack Aiki
add that directory to PATH
run aiki
```

No fixed installation prefix is required. No Aiki-specific environment variable identifies the installation. A release directory is relocatable.

`make dist` constructs that user distribution.

More importantly, `make distcheck` tests the distribution rather than merely testing the development tree. It unpacks Aiki under a temporary prefix, moves to an unrelated working directory, puts the unpacked Aiki on `PATH`, and proves that it can run a program using its shipped modules.

The test deliberately places distracting package trees in the unrelated working directory. They must not participate in named-module discovery.

There is also now a separate development artifact:

```text
make baseline
```

A baseline retains the repository, including Git state, so it can serve as a complete restart point. It is intentionally not the user distribution.

This distinction has turned out to matter repeatedly. A system that works only because it happens to be standing inside its own source tree has not established much.

## A Second Front End, Written in Aiki

The largest body of work since the alpha began with a more ambitious question:

Can Aiki describe Aiki independently?

There is now a second Aiki front end written in Aiki.

It includes an independent lexer, newline normalizer, and recursive-descent parser. It does not call through to the Go lexer or parser.

The independent implementation necessarily knows some of the same things as the production implementation. It must recognize Aiki keywords. It must recognize operators. It must understand the newline rule.

But duplicated implementation and duplicated authority are different things.

Where the second implementation has to copy enumerated facts, Aiki programs check those copies against facts derived from `grammar.ebnfx`. The algorithms remain independent.

The two front ends are then checked against reviewed neutral artifacts.

For lexical and newline behavior, both implementations must agree with common conformance projections.

For parsing, I did not invent a second parse-gold format. Aiki already had human-readable grammar-shaped parse golds used by the structural engine tests. The independent parser targets those same reviewed artifacts.

The relationship is therefore roughly:

```text
                 grammar / reviewed evidence
                         |
              +----------+----------+
              |                     |
              v                     v
        Go implementation      Aiki implementation
```

That is a much stronger arrangement than testing one implementation only against expectations constructed through that same implementation path.

## Language Knowledge as a Service

The self-description work and editor work turned out to share an architectural problem.

Aiki already knew most of the things an editor wants to ask:

* Is this source valid?
* What names are visible here?
* Where is this name defined?
* What does this operation mean?
* How should this document be formatted?
* What names are available for completion?

The wrong approach would have been to reimplement those answers in a VS Code extension or an LSP server.

So the important new component is not LSP.

It is the editor-independent language-service boundary beneath it.

Aiki now exposes reusable document analysis, diagnostics, symbols, definition lookup, canonical formatting, completion, and hover/help through that service layer.

`aiki lsp` is an adapter onto it.

That distinction has been useful because it immediately allowed several very different clients.

Xed uses its normal GtkSourceView machinery for lexical presentation and a thin plugin for live language-service diagnostics.

VS Code uses a conventional thin language client. The live path has been exercised for syntax recognition, diagnostics and their clearing, completion, hover, Go to Definition, and Format Document.

nvi does not need to pretend it is a modern IDE at all.

Aiki can generate traditional tags:

```text
aiki tags -o tags source.ai
```

and nvi uses them in the usual Unix fashion with `^]` and `^T`.

These are three very different interfaces to the same language.

The editor owns its buffer, cursor, presentation, and protocol mechanics.

Aiki owns Aiki.

Real editors also found real adapter defects.

Xed exposed Save-As document identity and zero-width EOF diagnostic presentation issues. Desktop launch made it clear that an editor's actual environment is not necessarily the environment of an interactive shell.

VS Code exposed a JSON-RPC mistake: a successful nullable response must contain an explicit `result: null`, not simply omit the result member. It also exposed an incorrect assumption that a definition request would arrive at the first character of an identifier rather than somewhere inside it.

None of these required editor-specific Aiki semantics.

That is the point.

## From Front End to Interpreter

Once the independent front end existed, I continued downward.

Aiki now contains an independent Aiki-written runtime environment and evaluator.

The environment supports lexical lookup, shadowing, assignment into the appropriate enclosing scope, closures, and shared capture using ordinary Aiki facilities.

The evaluator handles values, lists, shapes, expressions, left-to-right binary evaluation, indexing, field access, functions, calls, pipelines, statements, control flow, pattern matching, recursion, rest parameters, and returns.

Then self-hosting did exactly what I hoped it would do.

It found holes.

An interpreter reading:

```text
:foo
```

must be able to construct the corresponding native Aiki symbol dynamically.

Aiki had symbol syntax. It did not have a general operation that could turn the string `"foo"` into the symbol `:foo`.

I could have built a private symbol representation inside the second interpreter. I could have enumerated symbols. I could have smuggled source back through the Go evaluator to manufacture the value.

All of those would have weakened the claim.

Instead the language acquired the missing general capability:

```text
to_symbol("foo")
```

Likewise, the substrate could construct shaped lists internally, but ordinary Aiki lacked the corresponding general operation. That boundary was opened deliberately:

```text
shaped(:point, [1, 2])
```

These are small additions, but they matter to me because of why they exist.

Self-hosting demonstrated that the public value-construction surface was insufficient to express the language's own open syntax.

The language changed rather than the proof being weakened.

## Self-Hosting the Module Boundary

Modules exposed another boundary.

Delegating source-module loading to the host would have made the second evaluator much easier to complete, but it would also have delegated a large portion of language semantics back to the implementation being checked.

So Aiki-source module loading is now self-hosted.

The independent path performs the work itself:

```text
resolve
read
lex
normalize
parse
evaluate
collect exports
cache
```

Platform facilities remain platform facilities.

A privileged bootstrap captures the host capabilities that blessed Aiki library code legitimately requires. It provides those capabilities to self-hosted library environments without exposing the raw HAL surface to ordinary callers.

That is the boundary I want:

> Aiki source semantics can be implemented in Aiki. Host effects remain host effects.

Self-hosting does not require pretending the machine underneath has disappeared.

## Conformance Found a Semantic Error

The independent evaluator then began running existing behavior specimens.

That found a particularly useful mistake.

Aiki has recoverable error values such as:

```text
[@error, :math, "division by zero"]
```

They are data. They can be stored, returned, inspected, and pattern matched.

The second evaluator initially used Aiki's `is_error()` operation internally to recognize its own evaluation failures.

That was wrong.

It silently collapsed two different concepts: an Aiki error value and a failure of the interpreter itself.

The evaluator now uses a private self-interpreter fault representation for the latter. Recoverable Aiki errors remain ordinary Aiki values.

Focused implementation work had not exposed that distinction. Running broader behavior through an independent implementation did.

That is exactly why I wanted the second implementation.

## Aiki Interpreting Aiki Interpreting Aiki

The eventual target was full recursive self-interpretation.

The first serious attempt appeared to hit a wall while the outer Aiki-written interpreter was loading and parsing the inner parser.

The obvious explanation was parser recursion.

It was also wrong.

Aiki already had semantic profiling before the first public alpha. Here it became an engineering instrument.

Measurement showed that much of the cost was actually in character-level scanning. Repeated Aiki string indexing was repeatedly materializing rune views of large source strings. The supposedly parser-bound workload was dominated much earlier in the pipeline.

That realization led to a linear implementation behind the existing `string.chars` surface, along with changes to the self-host lexer so it snapshots characters once and avoids unnecessary table scans and repeated position work.

The measured workload improved substantially.

More importantly, once it ran far enough, it began exposing semantic problems instead of merely taking too long.

The nested module loader disagreed with the production loader about path fallback. Different spellings involving `.` and `..` could identify the same physical module as different cache entries. Later, running outside the source tree exposed another assumption: the self-hosted loader needed the same ordered knowledge of distribution module roots available to the production loader.

Those relationships were corrected.

No Go parser was substituted.

No Go AST was handed to the inner interpreter.

No host evaluator executed the target expression.

The resulting path now works:

```text
Go-hosted Aiki
    |
    v
Aiki-written interpreter
    |
    v
self-host-loaded Aiki interpreter
    |
    v
1 + 2 * 3
    |
    v
9
```

The `9` is useful because it carries one more piece of evidence.

At the third level, Aiki still evaluates the expression left to right:

```text
(1 + 2) * 3
```

The result is not simply "the interpreter ran." It preserves one of Aiki's deliberate and visible semantic commitments all the way down.

Go remains the production runtime and bootstrap substrate.

The self-hosted implementation is not an exercise in denying that dependency. It is an independent implementation and conformance boundary.

## Profiling the Profiler

Aiki already had semantic profiling before the first public alpha. During the self-hosting work, however, profiling stopped being merely an available facility and became an engineering instrument.

The first recursive runs appeared to spend most of their time in parser self-interpretation. Profiling showed that this diagnosis was wrong. Much of the cost was earlier in the pipeline, in repeated character and string work inside the independent lexer. That led to a linear realization behind `string.chars`, lexer snapshotting, and reductions in repeated scanning and dispatch work.

The measurements were useful.

They were also enormous.

That raised a different question:

> Why should I believe the profiler?

So before using those recursive measurements as evidence about Aiki, I calibrated the profiler itself.

The first calibration programs are deliberately trivial. The leaf computation:

```aiki
1 + 1
```

is placed inside one, two, three, and four nested loops of ten, causing it to execute:

```text
10
100
1,000
10,000
```

times.

At this scale the semantic work can be counted directly from the source.

For the four programs, the expected total loop-body iterations are:

```text
10
110
1,110
11,110
```

The profiler reports exactly those values.

Arithmetic is equally visible. Every leaf execution performs one addition and every loop-body iteration increments one loop variable, so:

```text
arithmetic = leaf calculations + loop iterations
```

The predicted arithmetic counts are therefore:

```text
20
210
2,110
21,110
```

The profiler reports exactly those values as well.

The comparison counts also agree with the source structure. Each loop condition is tested once more than its body executes, producing:

```text
11
121
1,221
12,221
```

Again, those are exactly the observed counts.

A simpler native expression provides another check:

```aiki
1 + 2 + 3 + 4
```

contains three arithmetic operations.

The profiler reports:

```text
arithmetic = 3
```

A conventional loop gives a still more explicit scaling relationship:

```aiki
while i < N {
    x = x + 1
    i = i + 1
}
```

For this program the expected counts are:

```text
arithmetic = 2N
comparison = N + 1
iteration  = N
```

For `N = 10`, `100`, and `1,000`, the profiler reports those relationships exactly.

So at the native level the profiler is not merely producing plausible numbers. Its semantic counters agree exactly with work that can be counted by inspection.

The experiment then moves outward into interpretation.

One Aiki-written interpreter is loaded and asked to evaluate the same small source expression zero, one, two, and four times.

The zero case establishes the fixed interpreter cost. Relative to that baseline, one additional interpreted evaluation contributes:

```text
arithmetic       958
comparison      1,082
call            3,255
iteration         832
index             935
```

Two evaluations contribute exactly twice those counts.

Four evaluations contribute exactly four times those counts.

The important result is not the absolute size of the numbers. It is the exact linearity of the added semantic work.

Then the experiment goes one level deeper.

Before any third-level probe is evaluated, merely reaching the second self-host level already produces:

```text
arithmetic       689,681
comparison     1,346,671
call           5,021,109
iteration        498,039
index            793,179
```

Now the counts are in the millions, but the test remains intentionally simple.

Evaluate the same third-level expression once.

Then evaluate it once more.

From zero to one evaluation, the profiler adds:

```text
arithmetic      +353,930
comparison      +388,932
call          +2,342,216
iteration       +304,091
index           +397,156
```

From one to two evaluations, it adds:

```text
arithmetic      +353,930
comparison      +388,932
call          +2,342,216
iteration       +304,091
index           +397,156
```

The two increment vectors are identical.

That is the strongest result in the calibration.

The same instrumentation that counts ten visible loop iterations exactly also accumulates identical repeated work when millions of semantic events separate the observations.

I ran the complete experiment twice.

Every semantic counter in the second run was identical to the corresponding counter in the first: native cases, one-level self-host cases, and two-level self-host cases.

The underlying Go realization figures were not invariant. Elapsed time, cumulative allocation traffic, allocation count, and garbage-collection cycles varied modestly between executions.

That distinction is important.

The semantic counters describe Aiki-level work.

Elapsed time and allocation behavior describe one realization of that work by the Go runtime on a particular machine.

Those are different kinds of measurement and should not be conflated.

The calibration therefore progresses from source whose work can be counted directly, through repeated interpretation, to recursive self-interpretation involving millions of semantic events.

At small scale, predicted and observed semantic counts agree exactly.

At one self-host level, repeated identical evaluations add exactly proportional work.

At two self-host levels, successive identical evaluations add identical semantic-count vectors.

That does not make wall-clock time, allocation traffic, malloc count, or garbage-collection activity invariant. Nor should it.

What it establishes is narrower and more useful:

> Large recursive profiler counts are not being trusted because they look plausible. They are produced by the same instrumentation whose behavior is exact on small, inspectable programs and additive under repeated interpreted work.

That gives the profiler an empirical calibration basis rather than merely an implementation pedigree.

## Experiments Are Not Tests

That calibration became **Experiment 001**, and in the process it exposed the need for another kind of project artifact.

Aiki now has a place for reproducible experiments.

This is deliberately separate from the test suite.

Tests answer questions such as:

> Does this required relationship still hold?

Experiments may ask:

> What happens when I do this?

or:

> Does this measurement behave as I expect across scale?

The experiment structure makes three things physically distinct:

```text
001-profiler-calibration/
    README.md

    experiment/
        PROCEDURE.md
        run.sh
        materials...

    results/
        observations...

    analyses/
        interpretations...
```

The root orients the reader.

`experiment/` says what was done, how, and with what expectations.

`results/` contains what actually happened.

`analyses/` contains what I infer from those observations.

That separation is intentional. Procedure is not result. Result is not interpretation. An interpretation should not quietly rewrite the procedure that produced the observation.

New experimental work can be started outside the source tree with:

```text
aiki experiment new "name"
```

Sequence identity comes from the running Aiki distribution's curated experiment collection; development occurs in the caller's working directory. Completed experiments can then be promoted into the repository and distribution.

This is a small facility, but it fits the project unusually well.

Aiki is a programming language, but it is also a research artifact. There should be a first-class place for empirical evidence that is neither documentation nor a correctness gold.

## The Commitments Have Not Changed

All of this has made Aiki considerably more elaborate internally.

It has not changed the commitments that motivated it.

I still want a small surface.

I still want computational relationships to be visible enough to apprehend rather than merely execute.

I still prefer direct composition over hidden machinery.

I still reject the idea that implementation accident should define the language.

And the use of generative AI has not become less important to the engineering argument as the implementation has grown. It has made the argument more important.

The answer is not to pretend AI did not produce implementation code.

The answer is to continue making more of the implementation answerable to things outside itself.

The grammar is an authority.

Reviewed behavior is an authority.

Executable documentation is an authority.

The public language surface is an authority.

Cross-implementation agreement is evidence.

A real editor exercising the language service is evidence.

A relocated distribution running independently of the source tree is evidence.

An experiment whose raw observations remain beside its procedure is evidence.

The direction is not toward trusting more machinery.

It is toward narrowing the places where machinery can make an unexamined claim.

## Platforms

Linux remains Aiki's development home.

That is where I build it, work on it, and exercise it continuously.

Intel macOS is periodically put through a more rigorous build and validation cycle rather than being treated as portable merely because the source looks portable. Windows is also run periodically and remains in the portability loop, although it does not receive the same continuous attention as Linux.

That asymmetry is intentional to state plainly.

Aiki is cross-platform alpha software developed on Linux. macOS and Windows are independently exercised, but Linux is where development lives.

## Pending Work and Proposals

Not everything exposed during this work has been resolved, and not every desirable facility belongs in the current milestone.

The grammar analysis, for example, made several newline-policy choices visible without changing them. Some expression continuations remain deliberately blocked by the current newline rule, while a function-literal ending exposes a different edge in the termination policy. Those are language-design questions now, not hidden parser behavior. I have left them that way until there is a reason to change the language rather than merely make the rule more symmetrical.

The language-service work likewise established the authority boundary without committing Aiki to every feature available through modern editor protocols. Incremental parsing, semantic tokens, rename/refactoring, workspace-wide indexing, and similar facilities remain candidates rather than assumed requirements. They should be added when the language needs them, and preferably after actual use demonstrates where the simpler full-analysis model becomes inadequate.

Spawned abnormal termination also deserves further attention. Aiki's isolation and message-passing model is strong precisely because spawned computation does not silently share an ordinary mutable environment with its caller. Failure and abnormal termination need equally clear observable semantics rather than host leakage or ad hoc exception behavior.

More generally, the repository now carries proposals, audit findings, decisions, experiments, and implementation sessions as different artifacts because they answer different questions. Proposed work should remain proposed until its evidence gate is satisfied. Deferred work should remain visible rather than becoming folklore.

## Future Direction

The continuing architectural direction is to tighten the noose around responsibilities and authorities: make each fact, capability, and policy belong somewhere definite; make consumers derive rather than duplicate; and make the relationships executable wherever they can be.

At the same time, I want to build out the systems substrate.

The question for each addition is not simply whether Aiki can expose another operating-system facility.

It is:

> What is the smallest thing the host must know, and what can Aiki construct from it?

### Process Execution

This is where I expect to begin.

A useful process facility needs to launch a program with arguments, standard input, environment, and working directory, then return standard output, standard error, and exit status.

The host responsibility is narrow. Process creation is an operating-system effect and belongs below the HAL boundary.

The useful orchestration belongs above it.

A small host capability can therefore unlock Aiki programs that compose ordinary external tools without putting process policy into the evaluator.

That immediately moves Aiki toward being able to do the sort of work for which I use Unix every day.

### Networking

TCP has a similarly clean boundary.

The host needs to provide the primitive effects: listen, accept, connect, read, write, close.

Aiki already has isolated `spawn`, channels, and message passing. Concurrent servers should emerge by composing networking with those existing facilities rather than by adding a special "server" subsystem.

That will also put useful pressure on the concurrency model.

### JSON

JSON probably does not need much host help at all.

The self-hosted lexer has already demonstrated character-level scanning in Aiki. The value model maps naturally to JSON's strings, numbers, booleans, arrays, objects, and null-like representation choices.

A JSON parser and emitter written principally in Aiki would be a useful library exercise precisely because it tests how much should remain above the HAL.

### Operating-System Surface

A systems language environment eventually needs ordinary process facts:

* arguments;
* environment variables;
* current working directory;
* directory change;
* process exit status;
* process ID;
* hostname;
* related small pieces of host state.

I expect this to be roughly a dozen or fifteen narrow host operations gathered behind an Aiki `os` library rather than scattered indiscriminately through the prelude.

The HAL should expose host facts.

The library should decide how Aiki programmers use them.

### Signals

Signals are more interesting because they meet the concurrency model.

SIGINT and SIGTERM should not become an unrelated asynchronous control mechanism that bypasses ordinary Aiki structure.

My current direction is that signal receipt should enter Aiki as an event delivered through a channel. A program can then coordinate it using the same message-passing and selection mechanisms it uses for other concurrent events.

### Timers

Timers should compose in the same fashion.

Wall-clock time, elapsed time, and sleeping are host facilities. Timed coordination belongs with concurrency.

`time.after` fits naturally with the existing `select` direction: produce a timed event and let ordinary Aiki composition decide what it means.

### Cryptographic Hashing

SHA-256 and HMAC are another relatively clear boundary.

They become necessary once programs begin doing real work involving integrity or authentication, and they can be exposed as narrow capabilities without teaching the evaluator anything about cryptographic protocols.

The recurring test will remain the same:

> Keep the host primitive small. Build the useful abstraction in Aiki.

## Aiki at `v0.4.0-alpha-26`

Much of the work above is architectural and therefore not immediately visible from a five-line REPL transcript.

For scale, it is useful to end with an inventory of what the tagged system actually contains.

### Language

* exact rational arithmetic;
* deliberate left-to-right ordinary binary evaluation;
* numbers, booleans, runes, strings, symbols, lists, and first-class functions;
* functions, closures, recursion, rest parameters, iteration, and pipelines;
* pattern matching;
* shaped list data and named fields;
* recoverable error values;
* modules and explicit imports;
* isolated spawn-based concurrency and channels;
* graphics, canvas, and turtle facilities.

### Language Authority and Description

* a declarative grammar owning the syntax surface;
* grammar-owned newline termination and suppression policy;
* centralized cached structural grammar analysis;
* explicit grammar/evaluator coupling;
* explicit grammar/formatter coupling;
* checked linter syntax knowledge;
* grammar-derived binary-operator membership;
* grammar-backed help for syntax policy;
* reviewed structural parse projections;
* executable documentation.

### Implementations

* the production Go-hosted Aiki interpreter;
* an independent Aiki-written lexer;
* an independent Aiki-written newline normalizer;
* an independent Aiki-written recursive-descent parser;
* an Aiki-written runtime environment;
* an Aiki-written evaluator;
* self-hosted Aiki-source module resolution, loading, exports, and caching;
* a scoped bootstrap bridge for legitimate HAL capabilities;
* complete recursive self-interpretation through a third-level Aiki program.

### Language Services

* editor-independent document analysis;
* structured lexical, parse, and lint diagnostics;
* symbol discovery;
* definition lookup;
* canonical formatting;
* completion from actual visible names;
* hover from source definitions and authored help;
* behavior-neutral observation and instrumentation;
* `aiki lsp` as a replaceable protocol adapter.

### Editors

* grammar-coupled Xed syntax support;
* live Xed language-service diagnostics;
* a thin VS Code LSP client;
* VS Code diagnostics, completion, hover, definition, and formatting;
* classic `aiki tags` generation for nvi and other tags consumers.

### Modules and Platform Boundary

* a standard Aiki module library;
* pure Aiki modules and HAL-backed modules;
* explicit distribution module roots;
* user library support;
* explicit relative path imports;
* relocatable distribution lookup based on the executable rather than accidental working-directory structure;
* dynamic native symbol construction with `to_symbol`;
* general shaped-value construction with `shaped`.

### Validation and Conformance

* Go unit and integration tests;
* Aiki-native tests;
* behavioral smoke tests;
* reviewed gold files;
* executable documentation examples;
* grammar-production coverage;
* structural engine golds;
* formatter parse-preservation checks;
* linter checks;
* tree integrity checking;
* fuzz tests;
* property tests;
* explicit negative-test declarations;
* cross-implementation lexer conformance;
* cross-implementation newline conformance;
* cross-implementation parser conformance;
* cross-implementation behavior checks;
* recursive self-interpretation invariants.

### Profiling

* deterministic Aiki-level semantic counts;
* arithmetic, comparison, call, iteration, indexing, channel, and store activity;
* optional source attribution;
* Go-runtime realization measurements including elapsed time and allocation activity;
* recursive self-host profiling;
* an empirical calibration path from inspectable native work through recursive interpretation.

### Experiments

* numbered reproducible experiment homes;
* executable procedures;
* retained raw observations;
* analyses kept distinct from results;
* distribution-aware experiment numbering;
* profiler calibration as Experiment 001.

### Distribution and Development

* `make validate`;
* `make dist`;
* `make distcheck`;
* restartable repository baselines;
* relocatable user distributions;
* structural checking of distributed artifacts;
* proposals for bounded architectural work;
* audit findings with explicit dispositions;
* recorded design decisions;
* the in-repository `ai/` engineering provenance and restart record.

## Where This Leaves Aiki

The objective remains the same: not merely to add capabilities, but to keep narrowing the places where an unstated assumption, duplicated fact, hidden host dependency, or convenient implementation accident can masquerade as part of the language.

The first public alpha said that Aiki had become substantial enough to put in front of other people.

`v0.4.0-alpha-26` says something different.

The language has now been made to explain considerably more of itself, to survive comparison with an independent implementation of itself, to expose its knowledge through several interfaces without surrendering authority to them, to run recursively through that second implementation, to work outside the shelter of its own source tree, and to subject its own measurements to calibration.

There is a great deal left to do.

That is the state I wanted to mark.

Development continues on `master`.

The tag is the place to stop and look.

### Links

* [Aiki on GitHub](https://github.com/decuser/aiki)
* [Aiki `v0.4.0-alpha-26`](https://github.com/decuser/aiki/tree/v0.4.0-alpha-26)
* [First Public Alpha](/posts/aiki-alpha-release/)

*post added 2026-08-15 19:38:00 -0500*


