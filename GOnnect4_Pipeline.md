# GOnnect4 — Pipeline CI/CD

Documentation de la pipeline GitHub Actions.

---

## 🎯 Fonctionnement

La pipeline se déclenche sur :
- **Push sur `main`** : Lint + Tests uniquement
- **Tag de version** (`1.0.0`, `2.1.3`, etc.) : Lint + Tests + Build Docker

---

## 📋 Jobs de la pipeline

### 1. Lint (toujours)

Vérifie la qualité du code Go :
- Server (`server/`)
- Client WASM (`client/wasm/`)

Linters activés :
- `errcheck` : Erreurs non gérées
- `govet` : Code suspect
- `staticcheck` : Analyse statique
- `unused` : Code inutilisé
- `misspell` : Fautes d'orthographe
- Etc.

### 2. Test (toujours)

Lance les tests unitaires :
- `server/lib/` : Tests de la logique métier

### 3. Build Docker (seulement sur tags)

Build et push l'image Docker avec plusieurs tags :
- `1.0.0` : Version exacte
- `1.0` : Dernière patch de 1.0.x
- `1` : Dernière minor de 1.x.x
- `latest` : Dernière version

---

## 🔄 Workflow

### Push normal (développement)

```bash
git add .
git commit -m "Add feature X"
git push origin main
```

**Résultat** :
- ✅ Lint
- ✅ Tests
- ❌ Pas de build Docker

### Release (version)

```bash
git tag 1.0.0 -m "Release 1.0.0"
git push origin 1.0.0
```

**Résultat** :
- ✅ Lint
- ✅ Tests
- ✅ Build Docker
- ✅ Push sur `ghcr.io/marvinegger/gonnect4:1.0.0` + `:latest`

---

## 🛠️ Linter en local

Avant de push, linter le code :

```bash
# Server + Client WASM
./scripts/lint.sh

# Seulement le server
cd server && golangci-lint run

# Seulement le client WASM
cd client/wasm && GOOS=js GOARCH=wasm golangci-lint run
```

---

## 📦 Versioning

Format **Semantic Versioning** : `MAJOR.MINOR.PATCH`

- `MAJOR` : Changements incompatibles (breaking changes)
- `MINOR` : Nouvelles fonctionnalités compatibles
- `PATCH` : Corrections de bugs

Exemples :
```bash
# Première version
git tag 1.0.0 -m "First stable release"

# Nouvelle fonctionnalité
git tag 1.1.0 -m "Add AI opponent"

# Correction de bug
git tag 1.1.1 -m "Fix AI crash"

# Breaking change
git tag 2.0.0 -m "Refactor game engine"
```

---

## 🔍 Vérifier la pipeline

### Sur GitHub

1. Aller sur le repo → Actions
2. Voir l'exécution de la pipeline
3. Vérifier que lint + test passent
4. Si tag : vérifier que le build Docker réussit

### Images Docker

Vérifier les images publiées :
- `https://github.com/marvinegger?tab=packages`

### En local

```bash
# Vérifier qu'une image existe
docker pull ghcr.io/marvinegger/gonnect4:1.0.0

# Lister les versions
docker images | grep gonnect4
```

---

## 🚨 Si la pipeline échoue

### Lint échoue

```bash
# Linter en local pour voir les erreurs
./scripts/lint.sh

# Corriger les erreurs
# Re-commit et push
```

### Tests échouent

```bash
# Lancer les tests en local
cd server/lib && go test -v .

# Corriger les bugs
# Re-commit et push
```

### Build Docker échoue

```bash
# Build en local pour débugger
docker build -t gonnect4:test .

# Voir les logs de la pipeline sur GitHub Actions
# Corriger le Dockerfile si nécessaire
```

---

## 🎯 Commandes utiles

```bash
# Linter avant commit
./scripts/lint.sh

# Tests avant commit
cd server/lib && go test -v .

# Créer un tag
git tag 1.0.0 -m "Release 1.0.0"
git push origin 1.0.0

# Voir les tags
git tag

# Supprimer un tag (si erreur)
git tag -d 1.0.0
git push origin --delete 1.0.0
```

---

## ✅ Checklist avant release

Avant de créer un tag de version :

- [ ] Code lint sans erreur (`./scripts/lint.sh`)
- [ ] Tests passent (`cd server/lib && go test -v .`)
- [ ] Fonctionnalités testées en local
- [ ] CHANGELOG.md mis à jour (si existe)
- [ ] Numéro de version correct (SemVer)

---

## 📊 Résumé

| Action | Lint | Test | Build Docker | Image publiée |
|--------|------|------|--------------|---------------|
| Push sur `main` | ✅ | ✅ | ❌ | ❌ |
| Tag `1.0.0` | ✅ | ✅ | ✅ | `1.0.0`, `1.0`, `1`, `latest` |

**Images** : `ghcr.io/marvinegger/gonnect4:<version>`
