/*
 * Produce a menu with light meals (caloric value < 10Kcal)
 *
 * Extracted from the presentation,
 *
 *   Inês Dutra
 *   "Constraint Logic Programming: a short tutorial"
 *   https://www.dcc.fc.up.pt/~ines/talks/clp-v1.pdf
 *   June 2010
 */

light_meal(A, M, D) :-
    I > 0, J > 0, K > 0, 
    I + J + K =< 10,
    starter(A, I),
    main_course(M, J),
    dessert(D, K).

meat(steak, 5).
meat(pork, 7).
fish(sole, 2).
fish(tuna, 4).
dessert(fruit, 2).
dessert(icecream, 6).

main_course(M, I) :-
    meat(M, I).
main_course(M, I) :-
    fish(M, I).

starter(salad, 1).
starter(soup, 6).
