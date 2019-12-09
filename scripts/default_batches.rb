require "json"

text = File.read("tracker.json")
json = JSON.parse(text)
batches = json["batches"]

base_url = "http://localhost:9001/bibutils/marc/?bib="

# Original batches in the tracker file.
batches.each_with_index do |b, i|
    start_range = "b" + b["chunk_start_bib"].to_s
    end_range = "b" + b["chunk_end_bib"].to_s
    filename = b["file_name"]
    puts "d.AddBatch(\"#{start_range}\", \"#{end_range}\", \"#{filename}\")"
end

# Combines two batches in the tracker file into a single one
# pairs = batches.count / 2
# (0..pairs-1).each do |i|
#     x = i * 2
#     y = x + 1
#
#     start_range = "b" + batches[x]["chunk_start_bib"].to_s
#     end_range = "b" + batches[y]["chunk_end_bib"].to_s
#     filename = "big_sierra_export_#{i}.mrc"
#
#     puts "d.AddBatch(\"#{start_range}\", \"#{end_range}\", \"#{filename}\")"
# end
#
# if (batches.count % 2) != 0
#     x = pairs * 2
#     start_range = "b" + batches[x]["chunk_start_bib"].to_s
#     end_range = "b" + batches[x]["chunk_end_bib"].to_s
#     filename = "big_sierra_export_#{x}.mrc"
#     puts "d.AddBatch(\"#{start_range}\", \"#{end_range}\", \"#{filename}\")"
# end
