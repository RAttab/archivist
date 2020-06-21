var archive = {
    path: {
        api: {
            query: function () {
                let url = new URL(document.URL);
                let guild = url.pathname.split("/")[2];
                return "/api/query/"+guild+url.search;
            },
            record: function (record) {
                let url = new URL(document.URL);
                let guild = url.pathname.split("/")[2];
                return "/api/record/"+guild+"/"+record;
            }
        },

        asset: {
            record: function (record) {
                let url = new URL(document.URL);
                let guild = url.pathname.split("/")[2];
                return "/asset/record/"+guild+"/"+record;
            }
        },

        page: {
            record: function (record) {
                let url = new URL(document.URL);
                let guild = url.pathname.split("/")[2];
                return "/record/"+guild+"/"+record;
            },
            gallery: function () {
                let url = new URL(document.URL);
                let guild = url.pathname.split("/")[2];
                return "/gallery/"+guild;
            }
        },

        discord: {
            icon: function () {
                // https://discord.com/branding
                return "https://discord.com/assets/f8389ca1a741a115313bede9ac02e2c0.svg";
            },
            msg: function (record) {
                return "https://discord.com/channels/" + record.guild + "/" + record.channel + "/" + record.message;
            }

        }
    },

    data: {
        tags: new Set()
    },

    gallery: function () {

        //TODO: Breaks the layout if I put it directly in the html page. Need to figure out why.
        let html = []
        html.push(`<div class="container">`);
        html.push(`<div class="sidebar"><div class="tags" /></div>`);
        html.push(`<div class="gallery"><div class="records" /></div>`);
        html.push(`</div>`);
        $("body").html(html.join(""));

        $.getJSON(archive.path.api.query(), archive.renderRecords);
    },

    renderRecords: function (records) {
        for (let record of records) {
            $("div.records").append(`<div id="`+record+`" class="record" />`);
            $.getJSON(archive.path.api.record(record), archive.renderRecord);
        };
    },

    renderRecord: function (record) {
        archive.renderTags(record.tags)
        archive.cursor = "img:"+record.img_id

        let html = [];
        html.push(`<a href="`+archive.path.page.record(record.id)+`">`);
        html.push(`<img src="`+archive.path.asset.record(record.id)+`">`);
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
        $("div.tags").html(html.join(""));
    },

    renderTag: function (tag) {
        let html = []
        html.push(`<div class="tag">`)
        html.push(`<a href="`+archive.path.page.gallery()+`?tag=`+encodeURIComponent(tag)+`">`+tag+`</a>`)
        html.push(`</div>`)
        return html.join("")
    }
}
