#!/bin/sh

# Basic Latin 0000–007F
# Latin-1 Supplement, 0080–00FF
# Latin Extended-A, 0100–017F
# Latin Extended-B, 0180–024F
# Greek Extended Unicode block U+1F00-1FFF
# 	"--unicodes=U+A,U+20,U+22,U+28,U+29,U+2C,U+2E,U+2F,U+32,U+33,U+3A,U+41-49,U+4C-4E,U+50,U+52,U+53,U+55-59,U+5B,U+5D,U+5F,U+61-70,U+72-7B,U+7D,U+A0,U+B7,U+3BB,U+3C4,U+20D7,U+2474,U+249C,U+24AD,U+24B8,U+24B9,U+24BC,U+24BE,U+24C1,U+2780-2784,U+278A-278E,U+28B1,U+FE5F,U+1F130,U+1F135,U+1F144,U+1F146",

# find uses via
#      % glyphhanger ../web/emb/frontpage.html

# bin/python3 -m venv venv
# source venv/bin/activate
# pip3 install fontTools
# pip install brotli

# too bad the results are so ugly... not quite sure why italic seems so janky

declare -a arr=("NotoSansDisplay-Bold" "NotoSansDisplay-Italic" "NotoSansDisplay-Thin"
"NotoSansDisplay_Condensed-SemiBold" "NotoSansMono_Condensed-Regular" "NotoSansDisplay-BoldItalic"
"NotoSansDisplay-Regular" "NotoSansDisplay_Condensed-Italic" "NotoSansDisplay_SemiCondensed-Italic"
"NotoSansDisplay-ExtraLight" "NotoSansDisplay-SemiBold" "NotoSansDisplay_Condensed-Regular"
"NotoSansDisplay_SemiCondensed-Regular")

for i in "${arr[@]}"
do
   pyftsubset ../web/emb/ttf/${i}.ttf \
   --unicodes=U+0000-024F,U+03FD,U+3BB,U+3C4,U+20D7,U+2474,U+249C,U+24AD,U+24B8,U+24B9,U+24BC,U+24BE,U+24C1,U+2780-2784,U+278A-278E,U+28B1,U+FE5F,U+1F00-1FFF \
   --output-file=./out/${i}Subset.ttf \
   --layout-features=*
done

declare -a arr=("FiraMono-Regular" "FiraSans-Bold" "FiraSans-BoldItalic"
"FiraSans-Italic" "FiraSans-Light" "FiraSans-Regular"
"FiraSans-SemiBold" "FiraSans-Thin" "FiraSansCondensed-Bold"
"FiraSansCondensed-Italic" "FiraSansCondensed-Regular")

for i in "${arr[@]}"
do
   pyftsubset ../web/emb/ttf/${i}.ttf \
   --unicodes=U+0000-024F,U+03FD,U+3BB,U+3C4,U+20D7,U+2474,U+249C,U+24AD,U+24B8,U+24B9,U+24BC,U+24BE,U+24C1,U+2780-2784,U+278A-278E,U+28B1,U+FE5F,U+1F00-1FFF \
   --output-file=./out/${i}Subset.ttf \
   --layout-features=*
done

#txt=`cat inuse.txt`
#for i in "${arr[@]}"
#do
#   pyftsubset ../web/emb/ttf/${i}.ttf \
#   --text= ${txt}\
#   --output-file=./out/${i}Subset.ttf \
#   --layout-features=* \
#   --ignore-missing-glyphs
#done