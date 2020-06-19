function renderRecords(records) {
    records.forEach(function (id, index) {
        var asset = "/asset/record/" + id;
        var body = $("body");
        body.append(`<div id="`+id+`"class="gallery"></div>`);

        $.getJSON("/api/record/"+id, function(record) {
            var html = []

            html.push(`<a target="_blank" href="`+asset+`">`+
                      `  <img src="`+asset+`" alt="%s" width="600" height="400">`+
                      `</a>`);
            html.push(`<div class="tags">`);

            record.tags.forEach(function (tag, index) {
                html.push(`<div class="tag">`);
                html.push(`<a href="/tag/`+encodeURIComponent(tag)+`">`+tag+`</a>`);
                html.push(`</div>`);
            });
            html.push(`</div>`);

            var limit = 90
            var caption = record.caption
            if (caption.length > limit) { caption = caption.substring(0, limit)+"..." }
            html.push(`<div class="desc">`+caption+`</div>`);

            $("div#"+record.id).html(html.join(""))
        })
    });
}
