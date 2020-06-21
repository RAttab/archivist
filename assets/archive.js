var archive = {
    url: {
        qs:     new URL(document.URL).search,
        guild:  new URL(document.URL).pathname.split("/")[2],
        record: new URL(document.URL).pathname.split("/")[3]
    },

    path: {
        api: {
            query:  function ()   { return "/api/query/" + archive.url.guild + archive.url.qs; },
            record: function (id) { return "/api/record/" + archive.url.guild + "/" + id; }
        },

        page: {
            gallery: function ()   { return "/gallery/" + archive.url.guild; },
            record:  function (id) { return "/record/"+ archive.url.guild + "/" + id; }
        },

        asset: {
            record: function (id) { return "/asset/record/" + archive.url.guild + "/" + id; }
        },

        discord: {
            icon: function () {
                // https://discord.com/branding
                return "https://discord.com/assets/f8389ca1a741a115313bede9ac02e2c0.svg";
            },
            record: function (record) {
                return "https://discord.com/channels/" + record.guild + "/" + record.channel + "/" + record.message;
            }
        }
    },

    data: {
        tags: new Set()
    },

    nav: function() {
        let html = [];
        html.push(`<a href="`+archive.path.page.gallery()+`">Top</a>`);
        $("#nav").html(html.join(""));
    },

    record: function () {
        $.getJSON(archive.path.api.record(archive.url.record), function (record) {
            archive.renderTags(record.tags)

            {
                let html = [];

                if (record.caption !== "") {
                    html.push(`<div id="caption">`+record.caption+`</div>`);
                }

                html.push(`<div id="image">`);
                html.push(`  <a href="`+archive.path.asset.record(record.id)+`">`);
                html.push(`    <img src="`+archive.path.asset.record(record.id)+`">`);
                html.push(`  </a>`);
                html.push(`</div>`);

                $("#view").html(html.join(""));
            }

            {
                let html = [];
                html.push(`<div id="discord" class="sidebar">`);
                html.push(`  <a href="`+archive.path.discord.record(record)+`">`);
                html.push(`    <img src="`+archive.path.discord.icon()+`">`);
                html.push(`    Open in discord`);
                html.push(`  </a>`);
                html.push(`</div>`);
                $("#sidebar").append(html.join(""));
            }

        });
    },

    gallery: function () {
        $.getJSON(archive.path.api.query(), archive.renderRecords);
    },

    renderRecords: function (records) {
        for (let record of records) {
            $("div#records").append(`<div id="`+record+`" class="record" />`);
            $.getJSON(archive.path.api.record(record), archive.renderRecord);
        };
    },

    renderRecord: function (record) {
        archive.renderTags(record.tags)
        archive.cursor = "img:"+record.img_id

        let html = [];
        html.push(`<a href="`+archive.path.page.record(record.id)+`">`);
        html.push(`  <img src="`+archive.path.asset.record(record.id)+`">`);
        html.push(`</a>`);
        $("div#"+record.id).html(html.join(""));
    },

    renderTags: function (tags) {
        let update = false;
        for (let tag of tags) {
            if (!archive.data.tags.has(tag)) {
                archive.data.tags.add(tag);
                update = true;
            }
        }
        if (!update) return;

        let html = []
        for (let tag of Array.from(archive.data.tags.keys()).sort()) {
            html.push(archive.renderTag(tag));
        }
        $("div#tags").html(html.join(""));
    },

    renderTag: function (tag) {
        let html = []
        html.push(`<div class="tag">`)
        html.push(`  <a href="`+archive.path.page.gallery()+`?tag=`+encodeURIComponent(tag)+`">`+tag+`</a>`)
        html.push(`</div>`)
        return html.join("")
    }
}
