for FILE in ~/data/marc_test/*.mrc
do
    marcli -file $FILE -field LDR | grep LDR | wc -l
done