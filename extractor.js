/*
copy the contents of this file, and drop it into the console of the browser tab (tested in Chrome)
a download will be triggered
note the location where that ____playst.json file is saved (~/Downloads)
*/
items = []
listTitle = document.getElementById('header').getElementsByClassName('ytmusic-detail-header-renderer')[3].getElementsByTagName('h2')[0].textContent.trim()
rows = document.getElementsByTagName('ytmusic-playlist-shelf-renderer')[0].getElementsByTagName('ytmusic-responsive-list-item-renderer')
for (row of rows) {
    items.push({
        'title':row.getElementsByTagName('yt-formatted-string')[0].attributes['title'].value,
        'artist': row.getElementsByTagName('yt-formatted-string')[1].attributes['title'].value,
        'album':row.getElementsByTagName('yt-formatted-string')[2].attributes['title'].value
    })
}
//console.log(items)
//https://stackoverflow.com/questions/19721439/download-json-object-as-a-file-from-browser
playlist = {
    'title': listTitle,
    'songs': items
}
var dataStr = "data:text/json;charset=utf-8," + encodeURIComponent(JSON.stringify(playlist));
var dlAnchorElem = document.createElement('a');
dlAnchorElem.setAttribute("href",     dataStr     );
dlAnchorElem.setAttribute("download", listTitle+"_playlist.json");
dlAnchorElem.click();
console.log("download triggered")