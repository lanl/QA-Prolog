QA Prolog
=========

[![Go Report Card](https://goreportcard.com/badge/github.com/lanl/QA-Prolog)](https://goreportcard.com/report/github.com/lanl/QA-Prolog)

Description
-----------

QA Prolog ("Quantum Annealing Prolog") compiles a subset of the [Prolog programming language](https://en.wikipedia.org/wiki/Prolog), enhanced with some native support for [constraint-logic programming](https://en.wikipedia.org/wiki/Constraint_logic_programming), into a [2-local Ising-model Hamiltonian function](https://en.wikipedia.org/wiki/Ising_model) suitable for execution on a [D‑Wave quantum annealer](https://www.dwavesys.com/).  Technically, though, QA Prolog produces a *classical* Hamiltonian function so it can in principle target classical annealers as well.

QA Prolog is largely a proof of concept, but it does hold out the possibility—not yet demonstrated—of improving Prolog program execution time by replacing backtracking with fully parallel annealing into a solution state.

Installation
------------

See [`INSTALL.md`](INSTALL.md).

Usage
-----

Run `qa-prolog --help` for a list of command-line options.  At a minimum, you'll need to provide `--query=`〈*Prolog goal*〉 and a filename corresponding to a database of Prolog facts and rules.

Here's an example (running on D‑Wave hardware):

```
$ qa-prolog --verbose --qmasm-args="-O1 --postproc=opt" --query='friends(P1, P2).' examples/friends.pl 
qa-prolog: INFO: Parsing examples/friends.pl as Prolog code
qa-prolog: INFO: Representing symbols with 3 bit(s) and integers with 1 bit(s)
qa-prolog: INFO: Storing intermediate files in /tmp/qap-227417173
qa-prolog: INFO: Writing Verilog code to friends.v
qa-prolog: INFO: Writing a Yosys synthesis script to friends.ys
qa-prolog: INFO: Converting Verilog code to an EDIF netlist
qa-prolog: INFO: Executing yosys -q friends.v friends.ys -b edif -o friends.edif
qa-prolog: INFO: Converting the EDIF netlist to QMASM code
qa-prolog: INFO: Executing edif2qmasm -o friends.qmasm friends.edif
qa-prolog: INFO: Executing qmasm --run --values=ints -O1 --postproc=opt --pin=Query.Valid := true friends.qmasm
P1 = charlie
P2 = alice

P1 = alice
P2 = charlie
```

Citation
--------

The following journal publication discusses the design and implementation of QA Prolog:

> Pakin, Scott. “Performing Fully Parallel Constraint Logic Programming on a Quantum Annealer.” [*Theory and Practice of Logic Programming*](https://www.cambridge.org/core/journals/theory-and-practice-of-logic-programming), [vol. 18, no. 5–6](https://www.cambridge.org/core/journals/theory-and-practice-of-logic-programming/issue/2E19771FEA173F5FEA03108D2142054C), pp. 928–949, September 2018.  Eds.: Ferdinando Fioretto and Enrico Pontelli.  Cambridge University Press. ISSN: [1475‑3081](https://www.cambridge.org/core/journals/theory-and-practice-of-logic-programming), DOI: [10.1017/S1471068418000066](https://dx.doi.org/10.1017/S1471068418000066), [arXiv:1804.00036 [cs.PL]](https://arxiv.org/abs/1804.00036).

License
-------

QA Prolog is provided under a BSD-ish license with a "modifications must be indicated" clause.  See [the LICENSE file](LICENSE.md) for the full text.

QA Prolog is part of the Hybrid Quantum-Classical Computing suite, known internally as LA-CC-16-032.

Author
------

Scott Pakin, <pakin@lanl.gov>
