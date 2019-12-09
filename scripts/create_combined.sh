#!/bin/sh
#
# Combines the small MARC files produced by Sierra into larger files.
# We compress these larger files before sharing them.
#
MARC_FILES_PATH=/Users/hectorcorrea/data/marc_no_toc
COMBINED_PATH=/Users/hectorcorrea/data/marc_test
FILE_COUNT=0
BATCH_COUNT=1
BATCH_SIZE=100

echo "Deleting previous files..."
rm $COMBINED_PATH/combined_*.mrc

for FILE in `find $MARC_FILES_PATH -name "*.mrc" | sort`
do

  FILE_SIZE=$(stat -f %z "$FILE")
  if [[ "$FILE_SIZE" == "0" ]]; then

    # Skip empty files
    echo "Skipping $FILE (empty)"

  else

    if [ "$FILE_COUNT" -eq "$BATCH_SIZE" ]; then
      BATCH_COUNT=$((BATCH_COUNT + 1))
      FILE_COUNT=1
    else
      FILE_COUNT=$((FILE_COUNT + 1))
    fi

    COMBINED_FILE="$COMBINED_PATH/combined_$BATCH_COUNT.mrc"

    echo "Processing $FILE"
    if [ "$FILE_COUNT" -eq 1 ]; then
      cat $FILE > $COMBINED_FILE
    else
      cat $FILE >> $COMBINED_FILE
    fi

  fi

done

echo "Done creating combined files"
