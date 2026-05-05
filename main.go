package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

const layout = `<!DOCTYPE html>
<html lang="{{.Lang}}">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no">
    <title>LiteForm</title>
    <style>
        :root {
            --primary: #007AFF;
            --bg: #F5F5F7;
            --card-bg: rgba(255, 255, 255, 0.8);
            --text: #1D1D1F;
            --text-secondary: #86868B;
            --border: rgba(0, 0, 0, 0.1);
            --radius: 16px;
        }

        * { box-sizing: border-box; -webkit-font-smoothing: antialiased; }

        body {
            font-family: -apple-system, BlinkMacSystemFont, "SF Pro Text", "Helvetica Neue", Arial, sans-serif;
            background-color: var(--bg);
            color: var(--text);
            margin: 0;
            display: flex;
            flex-direction: column;
            align-items: center;
            min-height: 100vh;
            padding: 20px;
        }

        .lang-switcher {
            align-self: flex-end;
            margin-bottom: 20px;
            font-size: 13px;
            font-weight: 500;
        }

        .lang-switcher a {
            color: var(--text-secondary);
            text-decoration: none;
            padding: 4px 8px;
            border-radius: 6px;
            transition: all 0.2s;
        }

        .lang-switcher a.active {
            color: var(--primary);
            background: rgba(0, 122, 255, 0.05);
        }

        .container {
            width: 100%;
            max-width: 800px;
            animation: fadeIn 0.6s cubic-bezier(0.23, 1, 0.32, 1);
            display: flex;
            flex-direction: column;
            align-items: center;
        }

        @keyframes fadeIn {
            from { opacity: 0; transform: translateY(10px); }
            to { opacity: 1; transform: translateY(0); }
        }

        .card {
            background: var(--card-bg);
            backdrop-filter: blur(20px);
            -webkit-backdrop-filter: blur(20px);
            border-radius: var(--radius);
            padding: 32px;
            box-shadow: 0 8px 32px rgba(0, 0, 0, 0.04);
            border: 1px solid rgba(255, 255, 255, 0.4);
            width: 100%;
            max-width: 540px;
            margin-bottom: 24px;
        }

        .card.wide {
            max-width: 100%;
        }

        h1, h2 {
            font-weight: 600;
            margin-top: 0;
            letter-spacing: -0.02em;
            text-align: center;
        }
        
        h3 {
            font-weight: 600;
            margin-top: 0;
            margin-bottom: 16px;
            letter-spacing: -0.01em;
            font-size: 18px;
        }

        p.subtitle {
            color: var(--text-secondary);
            text-align: center;
            margin-bottom: 32px;
        }

        form { display: flex; flex-direction: column; gap: 16px; }

        .field-group { display: flex; flex-direction: column; gap: 8px; }

        label {
            font-size: 14px;
            font-weight: 500;
            color: var(--text-secondary);
            margin-left: 4px;
        }

        input, textarea {
            font-family: inherit;
            font-size: 15px;
            padding: 12px 16px;
            border-radius: 10px;
            border: 1px solid var(--border);
            background: rgba(255, 255, 255, 0.5);
            transition: all 0.2s ease;
            outline: none;
        }

        input:focus {
            border-color: var(--primary);
            background: #fff;
            box-shadow: 0 0 0 4px rgba(0, 122, 255, 0.1);
        }

        button {
            font-family: inherit;
            font-size: 15px;
            font-weight: 600;
            padding: 12px 16px;
            border-radius: 10px;
            border: none;
            background: var(--primary);
            color: white;
            cursor: pointer;
            transition: all 0.2s ease;
        }

        button:hover {
            opacity: 0.9;
            transform: scale(1.01);
        }

        button:active {
            transform: scale(0.98);
        }

        .success-icon {
            font-size: 48px;
            color: #34C759;
            text-align: center;
            margin-bottom: 16px;
        }

        /* Admin Styles */
        .admin-nav {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 24px;
            width: 100%;
        }

        .copy-btn {
            color: var(--primary);
            font-weight: 600;
            cursor: pointer;
            font-size: 12px;
            padding: 4px 8px;
            border-radius: 6px;
            background: rgba(0, 122, 255, 0.05);
            display: inline-block;
        }

        .copy-btn:hover { background: rgba(0, 122, 255, 0.1); }

        .table-container {
            width: 100%;
            overflow-x: auto;
            border-radius: 12px;
            border: 1px solid var(--border);
            background: #fff;
        }

        table {
            width: 100%;
            border-collapse: collapse;
            font-size: 14px;
            white-space: nowrap;
        }

        th {
            background: #F9F9FB;
            text-align: left;
            padding: 12px 16px;
            border-bottom: 1px solid var(--border);
            color: var(--text-secondary);
            font-weight: 600;
        }

        td {
            padding: 12px 16px;
            border-bottom: 1px solid var(--border);
            color: var(--text);
            vertical-align: middle;
        }
        
        .inline-form {
            display: flex;
            gap: 8px;
            align-items: center;
            flex-direction: row;
        }

        @media (max-width: 480px) {
            .card { padding: 24px; }
            body { padding: 15px; }
        }
    </style>
</head>
<body>
    <div class="lang-switcher">
        <a href="?lang=zh" class="{{if eq .Lang "zh"}}active{{end}}">中文</a>
        <span style="color:var(--border)">|</span>
        <a href="?lang=en" class="{{if eq .Lang "en"}}active{{end}}">English</a>
    </div>
    <div class="container">
        {{.Content}}
    </div>
    <script>
        function copyText(elementId, btn, copiedText) {
            const urlEl = document.getElementById(elementId);
            const path = urlEl ? urlEl.getAttribute('href') : '';
            const url = window.location.origin + path;
            navigator.clipboard.writeText(url).then(() => {
                const original = btn.innerText;
                btn.innerText = copiedText;
                setTimeout(() => btn.innerText = original, 2000);
            });
        }

        function autoExpand(el) {
            el.style.height = 'auto';
            el.style.height = (el.scrollHeight) + 'px';
        }

        document.addEventListener('input', function (event) {
            if (event.target.tagName.toLowerCase() !== 'textarea') return;
            autoExpand(event.target);
        }, false);
    </script>
</body>
</html>`

type PageData struct {
	Lang    string
	Content template.HTML
}

var i18n = map[string]map[string]string{
	"zh": {
		"title":           "信息登记",
		"subtitle":        "请填写以下信息完成提交",
		"submit":          "立即提交",
		"success":         "提交成功",
		"done_msg":        "您的信息已安全送达。",
		"again":           "再次填写",
		"admin":           "管理后台",
		"export":          "导出 CSV",
		"copy":            "复制",
		"fields_zh":       "中文字段 (逗号分隔)",
		"fields_en":       "英文字段 (对应顺序)",
		"save":            "保存配置",
		"time":            "时间",
		"unauth":          "未授权",
		"login_msg":       "请使用管理员账号登录",
		"clear":           "清空所有数据",
		"confirm":         "确定要物理清空所有提交记录吗？此操作不可恢复！",
		"please_use_link": "请使用分享链接访问",
		"link_invalid":    "链接无效或不存在",
		"link_expired":    "链接已过期",
		"password_req":    "需要密码访问",
		"password_err":    "密码错误",
		"password":        "访问密码",
		"enter_pwd":       "请输入访问密码",
		"verify":          "验证",
		"link_gen":        "生成分享链接",
		"link_label":      "备注名称",
		"link_pwd":        "访问密码 (可选)",
		"link_exp":        "过期日期 (可选 YYYY-MM-DD)",
		"generate":        "生成链接",
		"links_list":      "分享链接管理",
		"link_url":        "访问链接",
		"no_pwd":          "无密码",
		"no_exp":          "永不过期",
		"delete":          "删除",
		"copied":          "已复制",
		"update":          "更新",
		"source_link":     "来源",
		"delete_confirm":  "确定要删除此分享链接吗？删除后该链接将失效。",
	},
	"en": {
		"title":           "Registration",
		"subtitle":        "Please fill in the information below",
		"submit":          "Submit Now",
		"success":         "Success",
		"done_msg":        "Your information has been securely recorded.",
		"again":           "Submit Again",
		"admin":           "Admin Panel",
		"export":          "Export CSV",
		"copy":            "Copy",
		"fields_zh":       "Chinese Fields (Comma separated)",
		"fields_en":       "English Fields (Corresponding order)",
		"save":            "Save Config",
		"time":            "Time",
		"unauth":          "Unauthorized",
		"login_msg":       "Please login as admin",
		"clear":           "Clear All Data",
		"confirm":         "Are you sure you want to PERMANENTLY clear all records? This cannot be undone!",
		"please_use_link": "Please use a share link to access",
		"link_invalid":    "Link is invalid or does not exist",
		"link_expired":    "Link has expired",
		"password_req":    "Password Required",
		"password_err":    "Incorrect Password",
		"password":        "Access Password",
		"enter_pwd":       "Please enter access password",
		"verify":          "Verify",
		"link_gen":        "Generate Share Link",
		"link_label":      "Label / Note",
		"link_pwd":        "Password (Optional)",
		"link_exp":        "Expiry Date (Optional YYYY-MM-DD)",
		"generate":        "Generate Link",
		"links_list":      "Share Links Management",
		"link_url":        "Link URL",
		"no_pwd":          "No Password",
		"no_exp":          "Never Expires",
		"delete":          "Delete",
		"copied":          "Copied",
		"update":          "Update",
		"source_link":     "Source",
		"delete_confirm":  "Are you sure you want to delete this link? It will become invalid.",
	},
}

func main() {
	if _, err := os.Stat("./data"); os.IsNotExist(err) {
		os.Mkdir("./data", 0755)
	}
	var err error
	db, err = sql.Open("sqlite3", "./data/liteform.db")
	if err != nil {
		log.Fatal(err)
	}

	db.Exec(`CREATE TABLE IF NOT EXISTS configs (key TEXT PRIMARY KEY, value TEXT)`)
	db.Exec(`CREATE TABLE IF NOT EXISTS submissions (id INTEGER PRIMARY KEY, content TEXT, label TEXT DEFAULT '', created_at DATETIME DEFAULT CURRENT_TIMESTAMP)`)
	db.Exec(`CREATE TABLE IF NOT EXISTS share_links (token TEXT PRIMARY KEY, label TEXT, password TEXT, expires_at TEXT, created_at DATETIME DEFAULT CURRENT_TIMESTAMP)`)
	
	// Ensure label column exists for older DBs
	db.Exec(`ALTER TABLE submissions ADD COLUMN label TEXT DEFAULT ''`)

	db.Exec(`INSERT OR IGNORE INTO configs (key, value) VALUES ('fields_zh', '姓名,邮箱,备注')`)
	db.Exec(`INSERT OR IGNORE INTO configs (key, value) VALUES ('fields_en', 'Name,Email,Note')`)

	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/f/", handleForm)
	http.HandleFunc("/admin", handleAdmin)
	http.HandleFunc("/admin/link/create", handleLinkCreate)
	http.HandleFunc("/admin/link/delete", handleLinkDelete)
	http.HandleFunc("/admin/link/update", handleLinkUpdate)
	http.HandleFunc("/export", handleExport)
	http.HandleFunc("/clear", handleClear)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8255"
	}
	log.Printf("LiteForm started on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func getLang(r *http.Request) string {
	lang := r.URL.Query().Get("lang")
	if lang == "" {
		cookie, err := r.Cookie("lang")
		if err == nil {
			lang = cookie.Value
		}
	}
	if lang != "en" {
		lang = "zh"
	}
	return lang
}

func render(w http.ResponseWriter, r *http.Request, content string) {
	lang := getLang(r)
	http.SetCookie(w, &http.Cookie{Name: "lang", Value: lang, Path: "/"})
	tmpl, _ := template.New("layout").Parse(layout)
	tmpl.Execute(w, PageData{Lang: lang, Content: template.HTML(content)})
}

func parseExpiry(s string) (string, error) {
	if s == "" {
		return "", nil
	}
	s = strings.ReplaceAll(s, ",", "-")
	s = strings.ReplaceAll(s, " ", "-")
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return "", err
	}
	t = t.Add(24*time.Hour - time.Second)
	return t.Format(time.RFC3339), nil
}

func authCheck(w http.ResponseWriter, r *http.Request) bool {
	u, p, ok := r.BasicAuth()
	adminUser := os.Getenv("ADMIN_USER")
	if adminUser == "" {
		adminUser = "admin"
	}
	adminPass := os.Getenv("ADMIN_PASS")

	if !ok || u != adminUser || p != adminPass {
		w.Header().Set("WWW-Authenticate", `Basic realm="Admin"`)
		w.WriteHeader(401)
		return false
	}
	return true
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	lang := getLang(r)
	t := i18n[lang]
	render(w, r, fmt.Sprintf(`<div class="card"><h2>%s</h2></div>`, t["please_use_link"]))
}

func handleForm(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimPrefix(r.URL.Path, "/f/")
	lang := getLang(r)
	t := i18n[lang]

	var label, pwd, expStr string
	err := db.QueryRow("SELECT label, password, expires_at FROM share_links WHERE token=?", token).Scan(&label, &pwd, &expStr)
	if err != nil {
		render(w, r, fmt.Sprintf(`<div class="card"><h2>%s</h2></div>`, t["link_invalid"]))
		return
	}

	if expStr != "" {
		exp, err := time.Parse(time.RFC3339, expStr)
		if err == nil && time.Now().After(exp) {
			render(w, r, fmt.Sprintf(`<div class="card"><h2>%s</h2></div>`, t["link_expired"]))
			return
		}
	}

	if pwd != "" {
		cookieName := "auth_" + token
		cookie, err := r.Cookie(cookieName)
		if err != nil || cookie.Value != "1" {
			if r.Method == "POST" && r.FormValue("pwd") != "" {
				if r.FormValue("pwd") == pwd {
					http.SetCookie(w, &http.Cookie{Name: cookieName, Value: "1", Path: "/f/" + token})
					http.Redirect(w, r, "/f/"+token, 303)
					return
				} else {
					render(w, r, fmt.Sprintf(`<div class="card"><h2>%s</h2><p style="color:red">%s</p><form method="POST"><input type="password" name="pwd" placeholder="%s"><button type="submit">%s</button></form></div>`, t["password_req"], t["password_err"], t["password"], t["verify"]))
					return
				}
			}
			render(w, r, fmt.Sprintf(`<div class="card"><h2>%s</h2><form method="POST"><input type="password" name="pwd" placeholder="%s"><button type="submit">%s</button></form></div>`, t["password_req"], t["password"], t["verify"]))
			return
		}
	}

	var fZh, fEn string
	db.QueryRow("SELECT value FROM configs WHERE key='fields_zh'").Scan(&fZh)
	db.QueryRow("SELECT value FROM configs WHERE key='fields_en'").Scan(&fEn)

	fieldsZh := strings.Split(fZh, ",")
	fieldsEn := strings.Split(fEn, ",")

	displayFields := fieldsZh
	if lang == "en" {
		displayFields = fieldsEn
	}

	if r.Method == "POST" {
		var res []string
		for _, f := range fieldsZh {
			res = append(res, r.FormValue(strings.TrimSpace(f)))
		}
		db.Exec("INSERT INTO submissions (content, label) VALUES (?, ?)", strings.Join(res, "|"), label)
		render(w, r, fmt.Sprintf(`<div class="card">
            <div class="success-icon">✓</div>
            <h2>%s</h2>
            <p class="subtitle">%s</p>
            <button onclick="location.href='/f/%s?lang=%s'" style="width:100%%">%s</button>
        </div>`, t["success"], t["done_msg"], token, lang, t["again"]))
		return
	}

	formHtml := fmt.Sprintf(`<div class="card">
        <h1>%s</h1>
        <p class="subtitle">%s</p>
        <form method="POST">`, t["title"], t["subtitle"])
	for i, f := range displayFields {
		f = strings.TrimSpace(f)
		key := strings.TrimSpace(fieldsZh[i])
		formHtml += fmt.Sprintf(`
            <div class="field-group">
                <label>%s</label>
                <textarea name="%s" placeholder="%s %s" required autocomplete="off" rows="1" style="resize:none; overflow:hidden; min-height:48px; line-height:1.4"></textarea>
            </div>`, f, key, lang_prompt(lang), f)
	}
	formHtml += fmt.Sprintf(`
            <button type="submit">%s</button>
        </form>
    </div>`, t["submit"])
	render(w, r, formHtml)
}

func lang_prompt(lang string) string {
	if lang == "en" {
		return "Enter"
	}
	return "请输入"
}

func handleAdmin(w http.ResponseWriter, r *http.Request) {
	lang := getLang(r)
	t := i18n[lang]

	if !authCheck(w, r) {
		render(w, r, fmt.Sprintf(`<div class="card"><h2>%s</h2><p class="subtitle">%s</p></div>`, t["unauth"], t["login_msg"]))
		return
	}

	if r.Method == "POST" {
		db.Exec("UPDATE configs SET value=? WHERE key='fields_zh'", r.FormValue("fields_zh"))
		db.Exec("UPDATE configs SET value=? WHERE key='fields_en'", r.FormValue("fields_en"))
		http.Redirect(w, r, "/admin", 303)
		return
	}

	var fZh, fEn string
	db.QueryRow("SELECT value FROM configs WHERE key='fields_zh'").Scan(&fZh)
	db.QueryRow("SELECT value FROM configs WHERE key='fields_en'").Scan(&fEn)

	fieldsZh := strings.Split(fZh, ",")
	fieldsEn := strings.Split(fEn, ",")

	displayHeaders := fieldsZh
	if lang == "en" {
		displayHeaders = fieldsEn
	}

	content := fmt.Sprintf(`
        <div class="admin-nav">
            <h2 style="margin:0">%s</h2>
            <div>
                <a href="/export" style="margin-right:16px;font-weight:600;color:var(--primary);text-decoration:none;">%s</a>
                <a href="javascript:if(confirm('%s'))location.href='/clear'" style="color:#FF3B30;font-weight:600;text-decoration:none;">%s</a>
            </div>
        </div>
        
        <div class="card wide" style="margin-bottom: 24px;">
            <form method="POST">
                <div class="field-group">
                    <label>%s</label>
                    <input name="fields_zh" value="%s">
                </div>
                <div class="field-group">
                    <label>%s</label>
                    <input name="fields_en" value="%s">
                </div>
                <button type="submit" style="background:#000">%s</button>
            </form>
        </div>
        
        <div class="card wide" style="margin-bottom: 24px;">
            <h3>%s</h3>
            <form method="POST" action="/admin/link/create">
                <div class="field-group">
                    <label>%s</label>
                    <input name="label" placeholder="%s" required>
                </div>
                <div class="field-group">
                    <label>%s</label>
                    <input name="password" placeholder="%s">
                </div>
                <div class="field-group">
                    <label>%s</label>
                    <input name="expires_at" placeholder="%s">
                </div>
                <button type="submit" style="background:#34C759">%s</button>
            </form>
        </div>

        <div class="card wide" style="margin-bottom: 24px; padding:0; overflow:hidden;">
            <div style="padding: 24px; border-bottom: 1px solid var(--border);">
                <h3 style="margin:0;">%s</h3>
            </div>
            <div class="table-container" style="border:none; border-radius:0;">
                <table>
                    <thead>
                        <tr>
                            <th>%s</th>
                            <th>%s</th>
                            <th>%s</th>
                            <th>%s</th>
                            <th>操作</th>
                        </tr>
                    </thead>
                    <tbody>`,
		t["admin"], t["export"], t["confirm"], t["clear"],
		t["fields_zh"], fZh,
		t["fields_en"], fEn,
		t["save"],
		t["link_gen"], t["link_label"], t["link_label"], t["link_pwd"], t["link_pwd"], t["link_exp"], t["link_exp"], t["generate"],
		t["links_list"], t["link_label"], t["link_url"], t["password"], t["link_exp"])

	lrows, _ := db.Query("SELECT token, label, password, expires_at FROM share_links ORDER BY created_at DESC")
	defer lrows.Close()
	for lrows.Next() {
		var token, label, pwd, exp string
		lrows.Scan(&token, &label, &pwd, &exp)

		expDisplay := exp
		if exp == "" {
			expDisplay = ""
		} else {
			pt, _ := time.Parse(time.RFC3339, exp)
			expDisplay = pt.Format("2006-01-02")
		}

		content += fmt.Sprintf(`<tr>
            <td>%s</td>
            <td>
                <div style="display:flex;align-items:center;gap:8px;">
                    <a href="/f/%s" target="_blank" id="link_%s" style="color:var(--primary);text-decoration:none;">/f/%s</a>
                    <span class="copy-btn" onclick="copyText('link_%s', this, '%s')">%s</span>
                </div>
            </td>
            <td colspan="3">
                <form method="POST" action="/admin/link/update" class="inline-form" style="margin:0;gap:8px;">
                    <input type="hidden" name="token" value="%s">
                    <input name="password" value="%s" style="width:120px;padding:8px;" placeholder="%s">
                    <input name="expires_at" value="%s" style="width:120px;padding:8px;" placeholder="YYYY-MM-DD">
                    <button type="submit" style="padding:8px 12px;background:var(--primary);margin:0;">%s</button>
                    <button type="button" onclick="if(confirm('%s')){document.getElementById('del_%s').submit();}" style="padding:8px 12px;background:#FF3B30;margin:0;">%s</button>
                </form>
                <form id="del_%s" method="POST" action="/admin/link/delete" style="display:none;">
                    <input type="hidden" name="token" value="%s">
                </form>
            </td>
        </tr>`,
			label, token, token, token, token, t["copied"], t["copy"],
			token, pwd, t["no_pwd"], expDisplay, t["update"], t["delete_confirm"], token, t["delete"], token, token)
	}

	content += `</tbody></table></div></div>`

	content += fmt.Sprintf(`
        <div class="card wide" style="padding:0; overflow:hidden;">
            <div class="table-container" style="border:none; border-radius:0;">
                <table>
                    <thead>
                        <tr>
                            <th style="color:var(--primary)">%s</th>`, t["source_link"])

	for _, f := range displayHeaders {
		content += "<th>" + strings.TrimSpace(f) + "</th>"
	}
	content += fmt.Sprintf("<th>%s</th></tr></thead><tbody>", t["time"])

	rows, _ := db.Query("SELECT content, label, created_at FROM submissions ORDER BY id DESC")
	defer rows.Close()
	for rows.Next() {
		var c, l, ts string
		rows.Scan(&c, &l, &ts)
		content += "<tr>"
		content += "<td style='color:var(--primary);font-weight:500;'>" + l + "</td>"
		vals := strings.Split(c, "|")
		for _, v := range vals {
			content += "<td>" + v + "</td>"
		}
		for i := len(vals); i < len(fieldsZh); i++ {
			content += "<td>-</td>"
		}
		content += "<td style='color:#86868B; font-size:12px'>" + ts + "</td></tr>"
	}
	content += "</tbody></table></div></div>"
	render(w, r, content)
}

func handleLinkCreate(w http.ResponseWriter, r *http.Request) {
	if !authCheck(w, r) {
		return
	}
	if r.Method == "POST" {
		label := r.FormValue("label")
		pwd := r.FormValue("password")
		exp, _ := parseExpiry(r.FormValue("expires_at"))

		b := make([]byte, 4)
		rand.Read(b)
		token := hex.EncodeToString(b)

		db.Exec("INSERT INTO share_links (token, label, password, expires_at) VALUES (?, ?, ?, ?)", token, label, pwd, exp)
	}
	http.Redirect(w, r, "/admin", 303)
}

func handleLinkUpdate(w http.ResponseWriter, r *http.Request) {
	if !authCheck(w, r) {
		return
	}
	if r.Method == "POST" {
		token := r.FormValue("token")
		pwd := r.FormValue("password")
		exp, err := parseExpiry(r.FormValue("expires_at"))
		if err == nil {
			db.Exec("UPDATE share_links SET password=?, expires_at=? WHERE token=?", pwd, exp, token)
		}
	}
	http.Redirect(w, r, "/admin", 303)
}

func handleLinkDelete(w http.ResponseWriter, r *http.Request) {
	if !authCheck(w, r) {
		return
	}
	if r.Method == "POST" {
		token := r.FormValue("token")
		db.Exec("DELETE FROM share_links WHERE token=?", token)
	}
	http.Redirect(w, r, "/admin", 303)
}

func handleClear(w http.ResponseWriter, r *http.Request) {
	if !authCheck(w, r) {
		return
	}
	db.Exec("DELETE FROM submissions")
	db.Exec("VACUUM")
	http.Redirect(w, r, "/admin", 303)
}

func handleExport(w http.ResponseWriter, r *http.Request) {
	if !authCheck(w, r) {
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment;filename=liteform_data.csv")
	w.Write([]byte{0xEF, 0xBB, 0xBF})

	writer := csv.NewWriter(w)

	var fZh string
	db.QueryRow("SELECT value FROM configs WHERE key='fields_zh'").Scan(&fZh)
	fields := strings.Split(fZh, ",")
	
	header := []string{"来源"}
	header = append(header, fields...)
	header = append(header, "提交时间")
	writer.Write(header)

	rows, _ := db.Query("SELECT content, label, created_at FROM submissions ORDER BY id DESC")
	defer rows.Close()
	for rows.Next() {
		var c, l, ts string
		rows.Scan(&c, &l, &ts)
		
		row := []string{l}
		row = append(row, strings.Split(c, "|")...)
		row = append(row, ts)
		writer.Write(row)
	}
	writer.Flush()
}
