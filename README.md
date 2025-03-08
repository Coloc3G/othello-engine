# Othello Engine

Un moteur de jeu Othello/Reversi avec différentes implémentations d'intelligence artificielle et un système d'entraînement génétique.

## À propos du projet

Ce projet implémente le jeu d'Othello (aussi connu sous le nom de Reversi) avec un moteur d'IA avancé permettant différentes stratégies d'évaluation. Le système comprend:

- Un moteur de jeu Othello complet avec toutes les règles
- Plusieurs fonctions d'évaluation pour l'IA
- Un algorithme Minimax avec élagage Alpha-Beta
- Un système d'apprentissage génétique pour optimiser les coefficients d'évaluation
- Un système de comparaison de performances entre différentes versions d'IA
- Une bibliothèque d'ouvertures standard d'Othello

## Le système d'IA

L'IA utilise un algorithme Minimax avec élagage Alpha-Beta pour rechercher le meilleur coup possible. Ce qui distingue les différentes versions d'IA sont les fonctions d'évaluation et leurs coefficients.

### Composants d'évaluation

L'IA évalue la position du jeu en tenant compte de plusieurs facteurs:

1. **Matériel** - Décompte des pièces sur le plateau
2. **Mobilité** - Nombre de coups disponibles pour chaque joueur
3. **Coins** - Contrôle des positions de coin, stratégiquement importantes
4. **Parité** - Avantage lié au nombre pair/impair de coups restants
5. **Stabilité** - Pièces qui ne peuvent plus être retournées
6. **Frontière** - Pièces adjacentes aux cases vides (généralement vulnérables)

### Phases de jeu

Les coefficients pour chaque composant d'évaluation sont ajustés selon trois phases de jeu:

- **Phase initiale** (0-19 pièces)
- **Milieu de partie** (20-58 pièces)
- **Fin de partie** (59-64 pièces)

### Versions de l'IA

Le projet comporte trois versions principales de l'IA, chacune avec des coefficients différents:

1. **V1** - Version de base avec des coefficients simples choisis arbitrairements
2. **V2** - Version améliorée avec des coefficients optimisés

## Comparaison des Performances

Le projet inclut un système de test pour comparer les performances des différentes versions d'IA. Vous pouvez générer des visualisations de ces performances en utilisant l'outil ci-dessous.

```bash
# Exemple d'utilisation
go run cmd/trainer/main.go --compare --compare-games 100 --compare-depth 5 --openings
```

### Résultats de Comparaison

Voici les résultats des comparaisons entre les différentes versions d'IA:

#### V1 vs V2

```
V1 Wins    [█████████████████████████                         ] 161 (32.2 %)
Draws      [██                                                ] 18  (3.6  %)
V2 Wins    [██████████████████████████████████████████████████] 321 (64.2 %)
```

## Visualisations

Pour visualiser les performances de chaque version d'IA, utilisez l'outil de visualisation:

```bash
go run cmd/visualization/main.go
```

Cela générera des histogrammes montrant les performances relatives des différentes versions.
