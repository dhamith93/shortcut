document.addEventListener('DOMContentLoaded', () => {
    const mainArea = document.getElementById('main-area');
    const hostIpSpan = document.getElementById('host-ip');
    const filePickerBtn = document.getElementById('file-picker-btn');
    const refreshBtn = document.getElementById('refresh-btn');
    const deviceName = window.prompt('Please enter an unique device name', '');
    const fileTable = document.getElementById('file-table');
    const prevents = (e) => e.preventDefault();

    if (deviceName && deviceName.length > 0) {
        document.getElementById('self-device-name').innerHTML = deviceName;
    } else {
        location.reload();
    }

    filePickerBtn.addEventListener('click', e => {
        const input = document.createElement('input');
        input.type = 'file';
        input.setAttribute('multiple', 'true');

        input.onchange = e => { 
            const files = e.target.files; 
            processFiles([...files]);
        }

        input.click();
    });

    refreshBtn.addEventListener('click', e => getFileList());

    ['dragenter', 'dragover', 'dragleave', 'drop'].forEach(evt => {
        mainArea.addEventListener(evt, prevents);
    });

    ['dragenter', 'dragover'].forEach(evt => {
        mainArea.addEventListener(evt, e => {
            mainArea.style.backgroundColor = '#DCDCDC';
        });
    });

    ['dragleave', 'drop'].forEach(evt => {
        mainArea.addEventListener(evt, e => {
            mainArea.style.backgroundColor = '#FFFFFF';
        });
    });

    mainArea.addEventListener('drop', e => {
        const files = e.dataTransfer.files;
        processFiles([...files]);
    });

    function processFiles(files) {
        files.forEach(file => {
            const formData = new FormData();            
            formData.append('file', file);
            formData.append('device-name', deviceName);
            const config = {
                headers: {
                    'Content-Type': 'multipart/form-data'
                },
                onUploadProgress: (progressEvent) => {
                    document.getElementById('upload-progressbar').style.display = 'block';
                    let percentCompleted = Math.round( (progressEvent.loaded * 100) / progressEvent.total );
                    document.getElementById('progress-completed').style.width =  percentCompleted + '%';
                }
            };
            axios.post('upload', formData, config).then((response) => {
                if (response.data) {
                    populateTable(fileTable, response.data);
                }
            }, (error) => {
                console.error(error);
                alert('Error uploading the file... ' + error);
            }).finally(() => {
                document.getElementById('upload-progressbar').style.display = 'none';
            });
        });
    }

    function getFileList() {
        axios.get('files').then((response) => {
            if (response.data) {
                populateTable(fileTable, response.data);
            }
        }, (error) => {
            console.error(error);
        });
    }

    function getMeta() {
        axios.get('meta').then((response) => {
            if (response.data) {
                hostIpSpan.innerHTML = response.data.Url;
            }
        }, (error) => {
            console.error(error);
        });
    }

    getMeta();
    getFileList();    
});