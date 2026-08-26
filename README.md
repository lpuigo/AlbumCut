# AlbumCut

Outil en ligne de commande pour découper un fichier MP3 contenant un album
complet (enregistrement continu) en plusieurs fichiers MP3 individuels, un
par piste, à partir d'un fichier texte listant les temps de départ de
chaque piste.

Le découpage se fait par copie de flux (`ffmpeg -c copy`), sans
réencodage, afin de préserver la qualité audio d'origine.

## Prérequis

- [`ffmpeg` et `ffprobe`](https://ffmpeg.org/download.html) installés et
  accessibles depuis le `PATH` de la ligne de commande.

## Compilation

```
go build -o albumcut.exe .
```

## Usage

```
albumcut [-end-padding=N] <album.mp3> <tracks.txt>
```

- `album.mp3` : le fichier MP3 de l'album complet.
- `tracks.txt` : le fichier texte listant les pistes (voir format
  ci-dessous).
- `-end-padding=N` (optionnel, défaut `0`) : nombre de secondes ajoutées à
  la fin de chaque piste, sauf la dernière, pour éviter une coupe trop
  courte en fin de piste. Les pistes générées se chevauchent alors
  légèrement avec le début de la piste suivante ; la fin ne dépasse
  jamais la durée réelle du fichier source. Le flag doit être placé
  avant les arguments positionnels.

Les fichiers de sortie sont créés dans le même répertoire que
`album.mp3`, nommés `NN. TitreAlbum_TitrePiste.mp3` (`NN` = numéro de
piste sur 2 chiffres, `TitreAlbum` = nom du fichier MP3 source sans
extension).

### Exemple

```
albumcut "D:\Musique\MonAlbum.mp3" "D:\Musique\tracks.txt"
```

produira par exemple :

```
D:\Musique\01. MonAlbum_My Pulse Got Lost in the Air Vents.mp3
D:\Musique\02. MonAlbum_Thunder in the Marrow.mp3
D:\Musique\03. MonAlbum_Exit Wound Waltz.mp3
```

## Format du fichier de timestamps

Une ligne par piste :

```
HH:MM:SS - Titre de la piste
```

ou, pour les albums de moins d'une heure :

```
MM:SS - Titre de la piste
```

Le titre est optionnel : si absent après le timestamp, la piste est
nommée `Untitled`. Les espaces du titre sont conservés tels quels (ils
se retrouvent dans le nom du fichier de sortie).

Exemple (`tracks.txt`) :

```
00:00:00 - My Pulse Got Lost in the Air Vents
00:05:07 - Thunder in the Marrow
00:09:57 - Exit Wound Waltz
00:14:26 - Untitled
```

Chaque piste est découpée du timestamp de son début jusqu'au timestamp
de la piste suivante (exclu) ; la dernière piste va jusqu'à la fin
réelle du fichier MP3 source.

Les timestamps doivent être strictement croissants. Le fichier est
entièrement validé avant tout découpage ; en cas d'erreur, le message
indique le numéro de ligne, le contenu fautif et la raison du rejet.

## Tests

```
go test ./...
```

Les tests d'intégration (découpage effectif via `ffmpeg`) sont ignorés
automatiquement si `ffmpeg`/`ffprobe` ne sont pas trouvés dans le
`PATH`.
