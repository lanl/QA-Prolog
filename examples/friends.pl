/* "The enemy of my enemy is my friend." */

hates(alice, bob).
hates(bob, charlie).

enemies(P, Q) :- hates(P, Q).
enemies(P, Q) :- hates(Q, P).

friends(A, B) :-
	enemies(A, X),
	enemies(X, B),
	A \= B.
