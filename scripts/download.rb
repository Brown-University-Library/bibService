# This program downloads from Sierra the JSON for the records indicated
# in a file produced with Traject. The downloaded files can be processed
# later with compare.rb
#
require "net/http"
require "json"

def http_get(url)
  uri = URI.parse(url)
  http = Net::HTTP.new(uri.host, uri.port)
  request = Net::HTTP::Get.new(uri.request_uri)
  response = http.request(request)
  response.body
end


last_downloaded = ""
file = "./data/traject_file.json"
tracker_file = "./data/last_downloaded.txt"
if File.exists?(tracker_file)
  last_downloaded = File.read(tracker_file)
end

File.foreach(file).with_index do |line, line_num|
  json1 = JSON.parse(line)
  id = json1["id"].first
  if id <= last_downloaded
    # puts "Skipping #{id}"
    next
  end
  puts "Downloading #{id} (#{line_num})"
  url = "http://localhost:9001/bibutils/bib/?bib=#{id}"
  sierraResponse = http_get(url)
  File.write("./data/#{id}.json", sierraResponse)
  File.write(tracker_file, id)
end
