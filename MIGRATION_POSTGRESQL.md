# Configuration PostgreSQL

## Migration de SQLite vers PostgreSQL

L'application a été migrée pour utiliser PostgreSQL au lieu de SQLite. Voici ce qui a changé :

### Modifications apportées

1. **Driver de base de données** : Remplacement de `github.com/glebarez/go-sqlite` par `github.com/lib/pq`
2. **Syntaxe SQL** : 
   - Placeholders changés de `?` à `$1, $2, $3...`
   - `INTEGER PRIMARY KEY` remplacé par `SERIAL PRIMARY KEY`
   - `rowid` remplacé par `ctid` pour PostgreSQL
3. **Configuration via variables d'environnement** : La connexion à la base de données utilise maintenant `DATABASE_URL`

### Configuration locale

1. **Installer PostgreSQL** :
   ```bash
   sudo apt-get install postgresql postgresql-contrib
   ```

2. **Créer la base de données** :
   ```bash
   sudo -u postgres psql
   CREATE DATABASE groupietracker;
   CREATE USER youruser WITH PASSWORD 'yourpassword';
   GRANT ALL PRIVILEGES ON DATABASE groupietracker TO youruser;
   \q
   ```

3. **Créer un fichier `.env`** basé sur `.env.example` :
   ```bash
   cp .env.example .env
   ```

4. **Modifier `.env`** avec vos informations :
   ```
   DATABASE_URL=postgres://youruser:yourpassword@localhost:5432/groupietracker?sslmode=disable
   ```

### Déploiement sur Scalingo

Sur Scalingo, la variable `DATABASE_URL` est automatiquement configurée lorsque vous ajoutez un addon PostgreSQL :

```bash
scalingo --app your-app addons-add postgresql postgresql-starter-1024
```

L'URL de la base de données sera automatiquement disponible dans `DATABASE_URL`.

### Exécution de l'application

```bash
# Charger les variables d'environnement
export $(cat .env | xargs)

# Compiler et exécuter
go build -o server .
./server
```

### Notes importantes

- Le fichier `.env` est ignoré par Git (voir `.gitignore`)
- Ne jamais commiter vos identifiants de base de données
- L'API Deezer utilisée est publique et ne nécessite pas de clé API
