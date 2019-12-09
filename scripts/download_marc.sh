# Runs on the terminal and downloads all the MARC records
# via the Sierra API using the batches indicated in
# (./josiag/downloadedBatches.go). The batches were defined
# from the tracker.json file that Birkin produces.
#
# It takes a very long time because we have to wait 15 minutes
# before continuing every 100 batches.
go build && ./bibService settings.json download



