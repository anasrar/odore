# odore

Odore is project for view and modified Sengoku Basara 2 Heros (PS2 game) develop and publish by Capcom.

## Index Character

Base on `zensh_`.

```
00: Maeda Toshiie
01: Date Masamune
02: Sanada Yukimura
03: Oda Nobunaga
04: Nōhime
05: Takeda Shingen
06: Uesugi Kenshin
07: Mori Ranmaru
08: Itsuki
09: Sarutobi Sasuke
10: Akechi Mitsuhide
11: Xavi
12: Shimazu Yoshihiro
13: Honda Tadakatsu
14: Matsu
15: Kasuga
16: Hojo Ujimasa
17: Mori Motonari
18: Tokugawa Ieyasu
19: Chosokabe Motochika
20: Imagawa Yoshimoto
21: Damī / Dummy / Placeholder
22: Maeda Keiji
23: Toyotomi Hideyoshi
24: Takenaka Hanbei
25: Miyamoto Musashi
26: Katakura Kojuro
27: Azai Nagamasa
28: Oichi
29: Kennyo Honganji
30: Fuma Kotaro
31: Matsunaga Hisahide
```

## File Name Pattern

```
armyNN.epk: compressed MDB 3D model with YZ2, untested
busho_NN_(N): T32 UI text, untested
cp_em(NNN): T32 UI text, untested
cp_ka: T32 UI text, untested
cp_mu(_NN): T32 UI text, untested
cp_nav(_NN): T32 UI map, untested // TODO: mapping with map name
cp_on(_N): T32 UI text, untested
cp_plNN_(_N): T32 UI text, untested
cp_tec(_N): T32 UI text, untested
cp_v(_N): T32 UI text, untested
face_(NNN).: T32 character, untested
face_p(_NN): T32 character, untested, suspect used on loading screen or character selection
forces_(_NN): T32 UI text, untested
kamon_NN(_NN): T32 clan logo, untested, suspect used on various place
m_NNN.t32: T32 stage, untested, suspect used on menu select stage
mame_NNN.: T32 stage, untested, suspect used on menu select stage
marmyNN.ep: compressed MDB 3D model with YZ2, untested
mis(_N): T32 UI mission stage, untested, suspect used on stage mission
missio(_NN): T32 UI mission stage, untested, suspect used on stage mission
moji8.t32: T32 UI font atlas, untested
msg_a(_NN): T32 UI general/soldier headshot, untested, suspect used on dialog message
msg_plN(_NN): T32 UI character headshot, untested, suspect used on dialog message
name_NNN.: T32 UI character name match with index character, untested
nowLoa(_N): T32 title, untested, suspect used on tournament select mode
plNN.pmk_y: compressed MTN animation with YZ2, untested
plNN.ppk_y: compressed MDB 3D model player with YZ2, untested
price.spt: item price list (uint32), untested
rNNN.t32: compressed T32 stage texture with YZ2, untested
result_(_NN): T32 stage, untested, suspect used on menu select stage
st(_NNN).t32: T32 UI text, untested, suspect used on menu
tais(_NN): T32 UI text, untested, suspect used on menu
taisho_(_NN): T32 UI text, untested, suspect used on menu
tate(_NNN).: T32 character, untested, suspect used on loading screen
teki_(_NNN).: T32 character headshot, untested
tutor(_N): T32 tutorial
zensh(_N): T32 character, used on extras
```
