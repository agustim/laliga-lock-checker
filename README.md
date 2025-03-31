# Laliga lock checker 

Script en Go per comprovar si una sèrie de dominis estan bloquejats i, si és necessari, provar-los a través d'una VPN. Els resultats es guarden en un fitxer CSV amb hora, estat i latència.

Recentment, *LaLiga* ha implementat mesures per combatre la pirateria de continguts esportius, sol·licitant el bloqueig d'adreces IP que, segons ells, distribueixen contingut il·legal. Aquestes accions han generat controvèrsia, ja que moltes de les IP bloquejades estan associades a *Cloudflare*, una empresa que proporciona serveis d'infraestructura i seguretat a nombrosos llocs web legítims. Com a resultat, diversos usuaris i empreses han experimentat interrupcions en els seus serveis, afectant negativament la seva operativa.

Per exemple, segons una [notícia publicada a elDiario.es](https://www.eldiario.es/tecnologia/cloudflare-lleva-tribunales-laliga-bloqueos-indiscriminados-pirateria_1_12065352.html), aquestes mesures han provocat que milers de pàgines web legítimes hagin estat afectades pels bloquejos, causant perjudicis econòmics i tècnics als seus propietaris.

A més, *Cloudflare* ha iniciat accions legals contra *LaLiga*, argumentant que els bloquejos són desproporcionats i afecten milions d'usuaris que intenten accedir a llocs web no relacionats amb la pirateria.

Aquest context subratlla la importància de disposar d'eines com *laliga-lock-checker* per monitoritzar i detectar possibles bloquejos de dominis, especialment per a aquells que depenen de serveis com *Cloudflare* per a la seva presència en línia.

## 🧹 Funcionalitats

- Llegeix dominis des d’un fitxer JSON (`sites.json`).
- Fa peticions HTTP i comprova si responen.
- Si no responen, activa una connexió VPN (WireGuard) i ho torna a provar.
- Escriu els resultats en un fitxer CSV: `hora,domini,estat,latencia_ms`.
- Permet configurar-ho per línia de comandes, variables d'entorn o fitxer `.env`.

---

## 📦 Instal·lació

```bash
git clone https://github.com/agustim/laliga-lock-checker.git
cd laliga-lock-checker
go mod tidy
```

Assegura't de tenir [Go](https://golang.org/dl/) instal·lat (versió 1.20 o superior recomanada).

---

## ⚙️ Exemple de `.env`

Crea un fitxer `.env` amb la configuració de la VPN:

```dotenv
INPUT_FILE=sites.json
OUTPUT_FILE=resultats.csv
VPN_INTERFACE=vpnwg0
PRIVATE_KEY=./privatekey
PUBLIC_KEY=publickey=
ENDPOINT=example.com:51820
VPN_ADDRESS=10.0.0.1/24
FWMARK=51820
```

---

## 🚀 Execució

```bash
go run main.go --debug
```

També pots sobreescriure qualsevol configuració amb flags:

```bash
go run main.go \
  --input=altres_sites.json \
  --output=log.csv \
  --vpn-interface=wg0 \
  --private-key=./vpn.key \
  --public-key=pubkey= \
  --endpoint=vpn.example.com:51820 \
  --vpn-address=10.0.0.2/24 \
  --fwmark=12345 \
  --debug
```

---

## 📁 Format dels fitxers

### `sites.json`

```json
[
  "example.com",
  "google.com",
  "nomésdomini.cat"
]
```

### `resultats.csv`

```
hora,domini,estat,latencia_ms
2025-03-31 15:42:00,example.com,not blocked,52
2025-03-31 15:42:05,domini.cat,blocked,142
```

---

## ✅ Requisits

- `wireguard-tools` instal·lat (`wg`, `ip`, etc.)
- Permisos de root o `sudo` per activar la VPN
- Clau privada i pública configurades

---

## 📄 Llicència

Aquest projecte està sota llicència MIT. Pots fer-ne ús lliurement.