async function clearElement(element){
    while (element.firstChild) {
        element.removeChild(element.lastChild);
    }
}

function populateTable(table, data) {
    if (data) {
        clearElement(table).then(() => {
            [...data].forEach(item => {
                let row = table.insertRow(-1);
                let cell1 = row.insertCell(-1);
                cell1.innerHTML = item.Device;
                item.Files.forEach(file => {
                    let row = table.insertRow(-1);
                    let cell1 = row.insertCell(-1);
                    cell1.classList.add('start');
                    let cell2 = row.insertCell(-1);
                    cell1.innerHTML = '<a href="/files/'+item.Device+'/'+file.Name+'" target="_blank" rel="noopener noreferrer" download>'+file.Name+'</a>';
                    let size = (file.Size > 1000000) ? convertTo(file.Size, 'B', 'M') + 'MB' : convertTo(file.Size, 'B', 'K') + 'KB'
                    cell2.innerHTML = size;
                });
            });
        });
    }
}

function buildCollectionItem(index, deviceName, content) {
    const template = `<li class="collection-item avatar">
    <i class="material-icons circle">devices</i>
    <span class="title">_device_name_</span>
    <p id="content-_index_">_content_</p>
    <a href="#!" data-copied-label="copied-_index_" class="secondary-content copy" data-clipboard-target="#content-_index_">
        <span id="copied-_index_" class="copyMessage" style="opacity: 0;">Copied!</span>
        <i class="material-icons">content_copy</i>
    </a>
</li>`;
    content = content.replace(/(?:\r\n|\r|\n)/g, '<br>');
    return template.replaceAll('_index_', index).replace('_device_name_', deviceName).replace('_content_', content);
}

let convertTo = (amount, unit, outUnit) => {
    let out = null;
    switch (unit) {
        case 'B':
            if (outUnit === 'M') {
                out = (amount / 1024) / 1024;
            } else if (outUnit === 'K') {
                out = amount / 1024;
            }
            break;
        case 'M':
            if (outUnit === 'G') {
                out = amount / 1024;
            } else if (outUnit === 'T') {
                out = (amount / 1024) / 1024;
            }
            break;
        case 'M':
            if (outUnit === 'G') {
                out = amount / 1024;
            } else if (outUnit === 'T') {
                out = (amount / 1024) / 1024;
            }
            break;
        case 'G':
            if (outUnit === 'M') {
                out = amount * 1024;
            } else if (outUnit === 'T') {
                out = amount / 1024;
            }
            break;
        case 'T':
            if (outUnit === 'M') {
                out = (amount * 1024) * 1024;
            } else if (outUnit === 'G') {
                out = amount * 1024;
            }
            break;
    }
    return Math.round(out);
}