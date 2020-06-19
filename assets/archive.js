function renderRecords(records) {
    records.forEach(function (id, index) {
        let body = $("body");
        body.append(`<div id="`+id+`"class="gallery"></div>`);

        $.getJSON("/api/record/"+id, function(record) {
            let html = [];
            let asset = "/asset/record/" + id;

            html.push(`<a target="_blank" href="/record/`+id+`">`+
                      `  <img src="/asset/record/`+id+`" alt="`+record.caption+`" width="600" height="400">`+
                      `</a>`);
            html.push(`<div class="tags">`);

            record.tags.forEach(function (tag, index) {
                html.push(`<div class="tag">`);
                html.push(`<a href="/tag/`+encodeURIComponent(tag)+`">`+tag+`</a>`);
                html.push(`</div>`);
            });
            html.push(`</div>`);

            let limit = 90;
            let caption = record.caption;
            if (caption.length > limit) { caption = caption.substring(0, limit)+"..." };
            html.push(`<div class="desc">`+caption+`</div>`);

            $("div#"+record.id).html(html.join(""));
        })
    });
}

// https://discord.com/branding
function discordIconUrl() {
    return "https://discord.com/assets/f8389ca1a741a115313bede9ac02e2c0.svg";
}

function discordMessageUrl(record) {
    return "https://discord.com/channels/" + record.guild + "/" + record.channel + "/" + record.message;
}
