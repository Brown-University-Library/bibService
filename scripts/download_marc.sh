# Runs on the terminal and downloads all the MARC records
# via the Sierra API using the batches indicated in
# (./josiag/downloadedBatches.go). The batches were defined
# from the tracker.json file that Birkin produces.
#
# It takes a very long time because we have to wait 15 minutes
# before continuing every 50-100 batches and there are 3900+
# batches in total.
go build && ./bibService settings.json download


# Once all the files have been downloaded combine them
# into 30-40 large files:
#
# create_combined.sh


# (optional) To convert our MARC files to MARCXML
# use IndexData's yaz-marcdump utility:
#
# yaz-marcdump -i marc -o marcxml $FILE > $FILE_XML


# Then join them into a single tar file and compress it,
# the resulting file is about 2GB.
#
# tar -czvf brown-2019-12-09.tar.gz *.mrc

