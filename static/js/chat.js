var lastMsgID = 0;

(function() {
    var msgs = document.querySelectorAll('.chat-msg');
    if (msgs.length > 0) {
        lastMsgID = parseInt(msgs[msgs.length - 1].dataset.id || '0');
    }
    poll();
})();

function poll() {
    var box = document.getElementById('chat-box');
    if (!box) return;
    var relID = box.dataset.rel || '0';

    fetch('/chat/poll?rel_id=' + relID + '&last_id=' + lastMsgID)
        .then(function(r) { return r.json(); })
        .then(function(msgs) {
            if (msgs && msgs.length > 0) {
                msgs.forEach(function(m) {
                    var mine = m.sender_id === userIDFromPage() ? 'msg-mine' : 'msg-theirs';
                    var div = document.createElement('div');
                    div.className = 'chat-msg ' + mine;
                    div.dataset.id = m.id;
                    var html = '<div class="msg-sender">' + escHtml(m.sender.nickname) + '</div>';
                    if (m.content) html += '<div class="msg-text">' + escHtml(m.content) + '</div>';
                    if (m.image_path) html += '<div class="msg-image"><img src="/uploads/' + escHtml(m.image_path) + '" loading="lazy" onclick="this.classList.toggle(\'zoomed\')"></div>';
                    html += '<div class="msg-time">' + (m.created_at || '').substring(5, 16).replace('T', ' ') + '</div>';
                    div.innerHTML = html;
                    box.appendChild(div);
                    lastMsgID = m.id;
                });
                box.scrollTop = box.scrollHeight;
            }
        })
        .catch(function() {})
        .finally(function() { setTimeout(poll, 4000); });
}

function sendMessage(relID) {
    var input = document.getElementById('chat-input');
    var fileInput = document.getElementById('chat-image');
    var content = input.value.trim();
    if (!content && !fileInput.files[0]) return false;

    var fd = new FormData();
    fd.append('rel_id', relID);
    fd.append('content', content);
    if (fileInput.files[0]) fd.append('image', fileInput.files[0]);

    var err = document.getElementById('chat-error');
    err.classList.add('hidden');

    fetch('/chat/send', {
        method: 'POST', body: fd,
        headers: {'X-Requested-With': 'XMLHttpRequest'}
    })
    .then(function(r) { return r.json(); })
    .then(function(d) {
        if (d.error) { err.textContent = d.error; err.classList.remove('hidden'); }
        else { input.value = ''; fileInput.value = ''; document.getElementById('img-label').textContent = '图片'; setTimeout(poll, 500); }
    })
    .catch(function() { err.textContent = '发送失败'; err.classList.remove('hidden'); });
    return false;
}

function userIDFromPage() { return 0; }

function escHtml(s) {
    if (!s) return '';
    return s.replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;');
}
