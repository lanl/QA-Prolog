/*
 * Solve the Riddle of the Potions from "Harry Potter and the
 * Philosopher's Stone" by J. K. Rowling.
 *
 * QA Prolog implementation by Scott Pakin <pakin@lanl.gov>.
 */

%! eq_bit(?A:atom, ?B:atom, ?Eq:Atom) is nondet.
%
% Eq is 1 if A and B are the same atom, 0 otherwise.
eq_bit(A, B, 0) :- atom(A), atom(B), A \= B.
eq_bit(A, B, 1) :- atom(A), atom(B), A = B.

%! cardinality_of(-Type:atom, ?A:atom, ?B:atom, ?C:atom, ?D:atom,
%!                ?E:atom, ?F:atom, ?G:atom, ?Value:int) is nondet.
%
% There are Value elements in [A, B, C, D, E, F, G] that match Type.
cardinality_of(Type, A, B, C, D, E, F, G, Value) :-
    eq_bit(Type, A, Aval),
    eq_bit(Type, B, Bval),
    eq_bit(Type, C, Cval),
    eq_bit(Type, D, Dval),
    eq_bit(Type, E, Eval),
    eq_bit(Type, F, Fval),
    eq_bit(Type, G, Gval),
    Aval + Bval + Cval + Dval + Eval + Fval + Gval = Value.

%! poison_before_wine(?A:atom, ?B:atom) is nondet.
%
% Given adjacent bottles, poison_before_wine is true unless wine
% does not have poison to its left.
poison_before_wine(_, Right) :- Right \= wine.
poison_before_wine(poison, wine).

%! bottles(?A:atom, ?B:atom, ?C:atom, ?D:atom, ?E:atom, ?F:atom, ?G:atom) is nondet.
%
% bottles solves the potions puzzle.
bottles(A, B, C, D, E, F, G) :-
    % "Danger lies before you, while safety lies behind,
    % Two of us will help you, whichever you would find,
    % One among us seven will let you move ahead,
    % Another will transport the drinker back instead,
    % Two among our number hold only nettle wine,
    % Three of us are killers, waiting hidden in line.
    % Choose, unless you wish to stay here for evermore,
    % To help you in your choice, we give you these clues four:"
    cardinality_of(forward,  A, B, C, D, E, F, G, 1),
    cardinality_of(backward, A, B, C, D, E, F, G, 1),
    cardinality_of(wine,     A, B, C, D, E, F, G, 2),
    cardinality_of(poison,   A, B, C, D, E, F, G, 3),

    % "First, however slyly the poison tries to hide
    % You will always find some on nettle wine's left side;"
    A \= wine,
    poison_before_wine(A, B),
    poison_before_wine(B, C),
    poison_before_wine(C, D),
    poison_before_wine(D, E),
    poison_before_wine(E, F),
    poison_before_wine(F, G),

    % "Second, different are those who stand at either end,"
    % "But if you would move onwards, neither is your friend;"
    A \= G,
    A \= forward,
    G \= forward,

    % "Third, as you see clearly, all are different size,
    % Neither dwarf nor giant holds death in their insides;"
    C \= poison,    % Assume smallest.
    F \= poison,    % Assume largest.

    % "Fourth, the second left and the second on the right
    % Are twins once you taste them, though different at first sight."
    B = F.
