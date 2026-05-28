var lastMsgID = 0;
var myUID = 0;

(function() {
    var box = document.getElementById('chat-box');
    if (!box) return;
    myUID = parseInt(box.dataset.user || '0');

    var msgs = box.querySelectorAll('.chat-msg');
    if (msgs.length > 0) {
        lastMsgID = parseInt(msgs[msgs.length - 1].dataset.id || '0');
    }
    box.scrollTop = box.scrollHeight;
    poll();
})();

function poll() {
    var box = document.getElementById('chat-box');
    if (!box) return;
    var relID = box.dataset.rel || '0';

    fetch('/chat/poll?rel_id=' + relID + '&last_id=' + lastMsgID)
        .then(function(r) { return r.json(); })
        .then(function(msgs) {
            if (!msgs || !msgs.length) return;
            msgs.forEach(function(m) {
                if (m.id <= lastMsgID) return;
                var cls = (m.sender_id === myUID) ? 'msg-mine' : 'msg-theirs';
                appendMsg(box, m, cls);
                lastMsgID = m.id;
            });
            box.scrollTop = box.scrollHeight;
        })
        .catch(function(e) { console.log('Poll error:', e); })
        .finally(function() { setTimeout(poll, 3000); });
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

    var btn = document.querySelector('#chat-form button[type=submit]');
    if (btn) btn.disabled = true;

    fetch('/chat/send', {
        method: 'POST', body: fd,
        headers: {'X-Requested-With': 'XMLHttpRequest'}
    })
    .then(function(r) { return r.json(); })
    .then(function(d) {
        if (btn) btn.disabled = false;
        if (d.error) {
            alert(d.error);
            return;
        }
        input.value = '';
        fileInput.value = '';
        var lbl = document.getElementById('img-label');
        if (lbl) lbl.textContent = '图片';

        // 移除空状态提示
        var empty = document.querySelector('#chat-box div[style]');
        if (empty && empty.textContent.indexOf('暂无消息') >= 0) empty.remove();

        // 立即显示
        var box = document.getElementById('chat-box');
        appendMsg(box, d, 'msg-mine');
        lastMsgID = d.id;
        box.scrollTop = box.scrollHeight;
    })
    .catch(function(e) {
        if (btn) btn.disabled = false;
        console.log('Send error:', e);
    });

    return false;
}

function appendMsg(box, m, cls) {
    var div = document.createElement('div');
    div.className = 'chat-msg ' + cls;
    div.dataset.id = m.id;

    var senderName = (m.sender && m.sender.nickname) ? m.sender.nickname : '';
    var html = '<div class="msg-sender">' + escHtml(senderName) + '</div>';
    if (m.content) html += '<div class="msg-text">' + escHtml(m.content) + '</div>';
    if (m.image_path) html += '<div class="msg-image"><img src="/uploads/' + escHtml(m.image_path) + '" loading="lazy" onclick="this.classList.toggle(\'zoomed\')"></div>';
    var t = (m.created_at || '').substring(5, 16).replace('T', ' ');
    html += '<div class="msg-time">' + t + '</div>';
    div.innerHTML = html;
    box.appendChild(div);
}

function escHtml(s) {
    if (!s) return '';
    return s.replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;');
}
