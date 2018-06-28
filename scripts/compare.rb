# This program compares the JSON in a file produced with Traject after
# processing a set of our MARC records from Sierra against the JSON that
# we are producing in the bibService.
#
# The general process is:
#   Read each record in traject_file.json
#   Download as JSON the corresponding record via the Sierra API
#   Make sure the JSON from the API matches the JSON in the Traject file.
#
# In order to prevent this program from downloading from Sierra the data
# for every record everytime the program is run there is a way to use a
# cached version of the Sierra data stored in a local file (search for
# "docFromFile" on this file.) See also download.rb
#
require "net/http"
require "json"

def equal_array(ar1, ar2)

  if !ar2.kind_of?(Array)
    return false
  end

  if ar1.length == 1 && ar1[0] == "" && ar2.length == 0
    # Consider [""] == []
    return true
  end

  # sort the arrays
  ar1.sort!
  ar2.sort!

  # compare their individual values
  ar1.each_with_index do |x, i|
    if !equal_value(x, ar2[i])
      return false
    end
  end
  return true
end

def equal_value(v1, v2)
  return true if v1 == v2
  return false if v2 == nil

  if v1.kind_of?(String) && v2.kind_of?(String)
    # forgive differences in trailing punctuation for now.
    len1 = v1.length
    len2 = v2.length
    case
    when len1 == len2
      return true if v1[0..len1-2] == v2[0..len2-2]
    when len1 > len2
      return true if v1[0..len1-2] == v2
    when len1 < len2
      return true if v1 == v2[0..len2-2]
    end
  end
  return false
end

def compare_one(id, json1, json2, show_equal, sequence)
  ignore_keys = [
    "updated_dt",       # very likely different
    "building_facet",   # traject has a bug when codes have spaces
    "author_facet",     # traject kept duplicates
    "oclc_t",           # Millenium MARC files had a different value under the "001"
    "isbn_t",           # Traject left a trailing ":"
    "region_facet",     # slight difference in punctuation
    "new_uniform_title_author_display", # minor punctuation differences (including encoding of ampersand)
    "uniform_title_author_display", # minor punctuation differences (including encoding of ampersand)
    "uniform_related_works_display", # minor punctuation differences (including encoding of ampersand)
    "text",             # pending
    "toc_display",      # pending
    "marc_display"      # pending
  ]

  id_shown = false
  timestamp = ""
  if json1['updated_dt'] != json2['updated_dt']
    timestamp = "(#{json1['updated_dt']} vs #{json2['updated_dt']})"
  end
  json1.keys.each do |key|
    if ignore_keys.include?(key)
      next
    end

    value1 = json1[key]
    value2 = json2[key]

    if value1.kind_of?(Array)
      equal = equal_array(value1, value2 || [])
    else
      equal = equal_value(value1, value2)
    end

    if !equal
      if !id_shown
        puts "#{id} #{sequence} #{timestamp}"
        id_shown = true
      end
      if value2 == nil
        puts "\t#{key} \t| (MISSING)"
      else
        puts "\t#{key} \t| #{value1} | #{value2}"
      end
    else
      if show_equal
        if !id_shown
          puts "#{id} #{sequence} #{timestamp}"
          id_shown = true
        end
        puts "\t#{key} \t [EQUAL]"
      end
    end
  end

end

def http_get(url)
  uri = URI.parse(url)
  http = Net::HTTP.new(uri.host, uri.port)
  request = Net::HTTP::Get.new(uri.request_uri)
  response = http.request(request)
  response.body
end

file = "./data/traject_file.json"
# file = "one.json"
# file = "sierra_3580.json"
show_equal = false
File.foreach(file).with_index do |line, line_num|
  json_traject = JSON.parse(line)
  id = json_traject["id"].first
  if id
    # Use this URL to compare with the live API
    # url = "http://localhost:9001/bibutils/solr/doc/?bib=#{id}"

    # Use this URL to compare against a cached file
    url = "http://localhost:9001/bibutils/solr/docFromFile/?bib=#{id}"

    sierraResp = http_get(url)
    begin
      json_sierra = JSON.parse(sierraResp)
    rescue JSON::ParserError
      json_sierra = {}
    end
    # if json1["updated_dt"] != json2["updated_dt"]
    #   next
    # end
    compare_one(id, json_traject, json_sierra, show_equal, line_num)
  else
    puts "Error processing line: #{line_num}"
  end
end

puts "Done."
