var archive = {

    guild: function () {
        return new URL(document.URL).pathname.split("/")[2]
    },

    page: function (op, id) {
        return "/" + op + "/" + archive.guild() + "/" + id;
    },

    api: function (op, id) {
        return "/api/" + op + "/" + archive.guild() + "/" + id;
    },

    apiFromPage: function () {
        return "/api" + new URL(document.URL).pathname
    },

    asset: function (id) {
        return "/asset/record/" + archive.guild() + "/" + id
    },

    renderRecords: function (records) {
        records.forEach(function (id, index) {
            let body = $("body");
            body.append(`<div id="`+id+`"class="gallery"></div>`);

            $.getJSON(archive.api("record", id), function(record) {
                let html = [];

                html.push(`<a href="`+archive.page("record", id)+`">`+
                          `  <img src="`+archive.asset(id)+`" alt="`+record.caption+`" width="600" height="400">`+
                          `</a>`);

                html.push(`<div class="tags">`);
                record.tags.forEach(function (tag, index) {
                    urlTag = encodeURIComponent(tag)
                    html.push(`<div class="tag">`);
                    html.push(`<a href="`+archive.page("tag", urlTag)+`">`+tag+`</a>`);
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
    },

    // https://discord.com/branding
    discordIconUrl: function () {
        return "https://discord.com/assets/f8389ca1a741a115313bede9ac02e2c0.svg";
    },

    discordMessageUrl: function (record) {
        return "https://discord.com/channels/" + record.guild + "/" + record.channel + "/" + record.message;
    }
}
