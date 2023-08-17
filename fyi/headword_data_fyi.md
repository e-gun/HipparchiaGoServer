## Distribution of headwords

* Only a few words are used many times.
* Many words are used only a few times.

### Greek headwords (nb: only counting `gr` tables, not `in`, etc.)

* mean:	`942.2`
* median:	`8.0`

|                        | Count | Percentage|
|------------------------|-------|---|
| total headwords        | 116310  | 100|
| hapax legomena         | 25019 | 21.5|
| between 2 and 10       | 42209| 36.3|
| between 11 and 100     | 28861| 24.8|
| between 101 and 1000   | 14549| 12.5|
| between 1001 and 10000 | 4738| 4.1|
|  more than 10001       | 933| 0.8|

![workers vs time](../gitimg/greek_headwords.png)


top 25:
```
hipparchiaDB=# SELECT entry_name,total_count from dictionary_headword_wordcounts where gr_count > 0 and entry_name ~ '[^a-z]' order by total_count desc limit 25;
 entry_name | total_count
------------+-------------
 ὁ          |    11355114
 καί        |     4506255
 τίϲ        |     2469259
 ἔδω        |     1972852
 δέ         |     1954170
 εἰμί       |     1869310
 δέω¹       |     1789454
 δεῖ        |     1759738
 δέομαι     |     1758537
 εἰϲ        |     1453398
 αὐτόϲ      |     1282478
 τιϲ        |     1056406
 οὗτοϲ      |      911397
 ἐν         |      896257
 γάροϲ      |      768593
 γάρον      |      768501
 γάρ        |      768155
 οὐ         |      716285
 μένω       |      663460
 μέν        |      633761
 τῷ         |      588789
 ἐγώ        |      566646
 ἡμόϲ       |      521478
 κατά       |      513083
 Ζεύϲ       |      513061
(25 rows)
```

### Latin headwords

* mean:	`292.2`
* median:	`11.0`

|                       | Count | Percentage|
|-----------------------|-------|---|
| total headwords       | 37594  | 100|
| hapax legomena        | 6908 | 18.3|
| between 2 and 10      | 12465| 33.1|
| between 11 and 100    | 10983| 29.2|
| between 101 and 1000  | 5693| 15.1|
| between 1001 and 10000 | 1423| 3.8|
| more than 10001       | 122| 0.3|

![workers vs time](../gitimg/latin_headwords.png)


top 25:

``` 
hipparchiaDB=# SELECT entry_name,total_count from dictionary_headword_wordcounts where lt_count > 0 and entry_name ~ '[a-z]' order by total_count desc limit 25;
 entry_name | total_count
------------+-------------
 qui¹       |      251752
 et         |      227327
 in         |      183802
 edo¹       |      159394
 is         |      132489
 sum¹       |      118288
 hic        |      100679
 non        |       96475
 ab         |       91768
 ut         |       76201
 Cos²       |       71431
 si         |       68898
 ad         |       68236
 cum        |       67434
 ex         |       65251
 a          |       65234
 eo¹        |       58138
 ego        |       53870
 quis¹      |       52628
 Eos        |       51588
 tu         |       51375
 dico²      |       48884
 ille       |       44216
 sed        |       44131
 de         |       42696
(25 rows)
```