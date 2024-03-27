# Tēzaurs terminālī

Nav pārlūka un vēlies izmantot Tēzauru? Tēzaurs terminālī. <br>
Nav logošanas sistēma? Tēzaurs terminālī. <br>
Nav interneta savienojuma? Tēzaurs te... ne gluži. <br>

Pagaidām šī programma izmanto [tezaurs.lv](https://tezaurs.lv/), lai iegūtu (noskrāpētu) datus, jo Tēzaura atvērtajiem datiem trūkst informācija par locījumiem un citām lietam.

![piemērs](/assets/piemers.gif)

Šis projekts vēl tikai top un tam trūkst daudzas lietas, kas ir īstajā Tēzaurā.

Mērķis ir izstrādāt Tēzauru terminālī ar pilnu tezaurs.lv funkcionalitāti, kā arī iespēju izmantot bez interneta savienojuma. 

Šis **NAV* oficiāls Tēzaura projekts, es nestrādāju LUMII.

# Kā ieinstalēt

Pagaidām strādā tikai uz Linux un visticamāk arī macOS (neesmu testējis).
Ja vēlies izmantot uz Windows, tad nāksies instalēt WSL2.

### Nepieciešamās lietas:
- [go](https://go.dev/doc/install) (1.21.4+)
- [fzf](https://github.com/junegunn/fzf)

**Piezīme**: kaut kad nākotnē izlaidīšu gatavus binārijus, tad Go nebūs vajadzīgs, bet pagaidām jābūvē pašam.

Palaid šīs komandas, lai ieinstalētu programmu
```sh
git clone https://github.com/deimoss123/tezaurs-tui
cd tezaurs-tui
./install.sh
```

Šis palaidīs instalēšanas skriptu, kas vienkārši uzbūvē projektu ar Go un novieto programmu vietā, kur tā ir palaižama.

Tēzauru var palaist ar komandu `tezaurs`.

Ja komanda netiek atrasta, tad `~/.local/bin` nav pievienots PATH.
To var izdarīt pievienojot sekojošo rindiņu savas čaulas konfigurācijas failam (`~/.bashrc`, `~/.zshrc`).

```sh
export PATH="$PATH:$HOME/.local/bin"
```

# Kā lietot

Palaižot `tezaurs` bez papildus argumentiem tiks atvērts vārdu meklētājs (`fzf`).

## Taustiņi

Meklētājā:
- **↑** / **↓** - Naviģēt starp vārdiem sarakstā
- **Enter** - Izvēlēties vārdu
- **Esc** - Iziet

Saskarnē:
- **↑** / **k** - Patīt uz augšu par vienu rindu
- **↓** / **j** - Patīt uz leju par vienu rindu
- **u** - Patīt uz augšu par pusekrānu
- **d** - Patīt uz leju par pusekrānu
- **m** - Meklēt jaunu vārdu
- **Esc** / **q** / **ctrl+c** - Iziet

## Papildus komandas argumenti

### `tezaurs -h`

Izvada lietošanas pamācību.

### `tezaurs -t <vārds>`

Izvadīs terminālī tikai tekstu, bez saskarnes.
Tiks izvadīts tikai vārda skaidrojums, bez locījumiem.

![](/assets/piemers2.png)
