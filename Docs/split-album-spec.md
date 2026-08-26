# Outil de découpage d'album MP3 en pistes individuelles

## Contexte

Outil en ligne de commande pour découper un fichier MP3 contenant un album
complet (enregistrement continu) en plusieurs fichiers MP3 individuels,
un par piste. Les temps de départ de chaque piste sont décrits dans un
fichier texte annexe.

## Besoin fonctionnel

### Entrées

- **Fichier MP3 album** : chemin fourni en argument de la ligne de commande.
- **Fichier texte de timestamps** : chemin fourni en argument de la ligne
  de commande. Format attendu, une ligne par piste :

  ```
  HH:MM:SS - Titre de la piste
  ```

  ou, pour les albums de moins d'une heure :

  ```
  MM:SS - Titre de la piste
  ```

  Le titre est optionnel : si absent après le timestamp, la piste est
  nommée `Untitled`.

  Exemple de fichier valide (`tracks.txt`) :

  ```
  00:00:00 - My Pulse Got Lost in the Air Vents
  00:05:07 - Thunder in the Marrow
  00:09:57 - Exit Wound Waltz
  00:14:26 - Untitled
  ...
  ```

### Traitement

- Chaque piste est découpée du timestamp de son début jusqu'au timestamp
  de la piste suivante (exclu), la dernière piste allant jusqu'à la fin
  réelle du fichier MP3 source.
- Découpage **sans réencodage** quand c'est possible (copie de flux via
  `ffmpeg -c copy`), afin de préserver la qualité audio d'origine et
  d'éviter tout cycle de décompression/recompression inutile.
  - Si la copie sans réencodage échoue ou produit une coupure trop
    imprécise en tête de piste (limite connue de `-c copy` sur MP3,
    alignement sur les frames), prévoir un mode de repli avec
    réencodage léger, et le signaler clairement à l'utilisateur dans les
    logs de sortie.

### Sortie

- Fichiers nommés selon le format :

  ```
  NN. TitreAlbum_TitrePiste.mp3
  ```

  - `NN` : numéro de piste, toujours sur 2 chiffres (`01`, `02`, ... `10`, `11`...)
  - `TitreAlbum` : nom du fichier MP3 source, sans extension
  - `TitrePiste` : titre de la piste tel qu'il apparaît dans le fichier
    texte, **espaces conservés tels quels**

  Exemple : `01. MonAlbum_My Pulse Got Lost in the Air Vents.mp3`

- Les fichiers sont créés **dans le même répertoire** que le fichier MP3
  source.
- Pas de détection de collision de noms de fichiers : le numéro de piste
  (`NN`) garantit l'unicité par construction.

### Hors périmètre

- Pas de gestion des métadonnées ID3 (titre, artiste, album, numéro de
  piste dans les tags) sur les fichiers de sortie.
- Pas d'interface graphique dans une première version (à réévaluer plus
  tard si besoin).

### Gestion des erreurs / validation

Le fichier de timestamps doit être validé avant tout découpage, avec des
messages d'erreur clairs, précis et actionnables, incluant si possible :
- le numéro de ligne fautif
- le contenu de la ligne fautive
- la raison précise du rejet

Cas à couvrir :
- Ligne mal formée / timestamp illisible
- Timestamps non strictement croissants (doublon ou désordre)
- Fichier de timestamps vide ou sans ligne valide
- Dernier timestamp situé après la fin réelle du fichier MP3 (incohérence
  détectée via la durée totale du fichier)
- Fichier MP3 introuvable
- `ffmpeg` (et `ffprobe`) non installé ou introuvable dans le `PATH`

## Contraintes techniques

- **Langage : Go**, natif, en minimisant au maximum les dépendances
  externes au langage lui-même (pas de librairies tierces si possible).
- **Dépendance externe autorisée** : `ffmpeg` / `ffprobe`, doivent être
  installés sur la machine et appelés via `exec.Command`.
- **Interface** : ligne de commande uniquement pour cette première
  version.
- **Environnement d'exécution** : outil lancé en local sur la machine de
  l'utilisateur (pas d'exécution distante).

## Plan d'exécution

### Étape 1 — Squelette du projet et CLI
- Initialisation du module Go (`go mod init`)
- Parsing des arguments en ligne de commande : chemin du MP3, chemin du
  fichier texte de timestamps
- Vérification de présence des fichiers, et vérification que `ffmpeg`
  (et `ffprobe`) sont installés et accessibles dans le `PATH` (message
  d'erreur explicite sinon, avec suggestion d'installation)

### Étape 2 — Parsing et validation du fichier de timestamps
- Parser chaque ligne selon une règle tolérante : `(HH:MM:SS|MM:SS) - Titre`
  (titre optionnel)
- Convertir chaque timestamp en secondes
- Appliquer les règles de validation décrites ci-dessus, avec message
  d'erreur précis incluant numéro de ligne et contenu de la ligne
  fautive
- Titre manquant après le timestamp → remplacé par `"Untitled"`

### Étape 3 — Calcul des segments de découpe
- Pour chaque piste : `début = timestamp[i]`, `fin = timestamp[i+1]`
  (ou durée totale du MP3 pour la dernière piste)
- Récupération de la durée totale du MP3 via `ffprobe` pour calculer la
  durée de la dernière piste et détecter une incohérence (dernier
  timestamp situé après la fin réelle du fichier → erreur claire)

### Étape 4 — Génération des noms de fichiers de sortie
- Construction du nom : `NN. TitreAlbum_TitrePiste.mp3`, `NN` sur 2
  chiffres, `TitreAlbum` = nom du fichier MP3 sans extension
- Sanitization minimale du titre pour rester un nom de fichier valide
  sur le système (échapper `/`, `\`, `:`, etc.), sans toucher aux
  espaces
- (Pas de détection de collision nécessaire : le numéro de piste
  garantit l'unicité)

### Étape 5 — Découpage effectif via ffmpeg
- Appel `ffmpeg -i album.mp3 -ss <début> -to <fin> -c copy <sortie>.mp3`
  pour chaque piste (copie de flux, pas de réencodage)
- Mode de repli avec réencodage léger si la copie sans réencodage échoue
  ou produit une coupure trop imprécise, signalé clairement à
  l'utilisateur
- Capture et affichage des erreurs ffmpeg de façon lisible (pas
  seulement le code retour brut)

### Étape 6 — Rapport final
- Résumé en fin d'exécution : nombre de pistes créées, chemins des
  fichiers, éventuels avertissements (ex: piste très courte, repli sur
  réencodage)

### Étape 7 — Tests
- Test avec un fichier de timestamps représentatif (mix de formats
  `HH:MM:SS` et `MM:SS`, présence d'une piste sans titre → `Untitled`)
- Tests de cas d'erreur : fichier timestamps vide, timestamps
  désordonnés, `ffmpeg` absent, MP3 introuvable

### Étape 8 — Packaging
- Compilation d'un binaire unique (`go build`)
- Court `README` expliquant l'usage, le format attendu du fichier de
  timestamps, et la nécessité d'avoir `ffmpeg` installé sur la machine
