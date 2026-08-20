package command

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"filvault/cli/internal/client"
)

// Commands holds shared dependencies for CLI subcommands.
type Commands struct {
	Client          *client.Client
	Out             io.Writer
	Err             io.Writer
	CurrentFolderID string
	SaveFolder      func(id string) error
}

// Login authenticates and persists tokens via the provided saver.
func (c *Commands) Login(email, password string, save func(access, refresh string) error) error {
	var session struct {
		User         user   `json:"user"`
		AccessToken  string `json:"accessToken"`
		RefreshToken string `json:"refreshToken"`
	}
	if err := c.Client.Do(http.MethodPost, "/auth/login", map[string]string{
		"email":    email,
		"password": password,
	}, &session); err != nil {
		return err
	}
	if save != nil {
		if err := save(session.AccessToken, session.RefreshToken); err != nil {
			return err
		}
	}
	fmt.Fprintf(c.Out, "Logged in as %s\n", email)
	return nil
}

type user struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	DisplayName  string `json:"displayName"`
	StorageUsed  int64  `json:"storageUsed"`
	StorageQuota int64  `json:"storageQuota"`
}

type folder struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	ParentID *string `json:"parentId"`
}

type file struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	SizeBytes int64  `json:"sizeBytes"`
}

type browser struct {
	Folder  *folder  `json:"folder"`
	Folders []folder `json:"folders"`
	Files   []file   `json:"files"`
}

type trashItem struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	SizeBytes *int64 `json:"sizeBytes"`
}

type trashList struct {
	Folders []trashItem `json:"folders"`
	Files   []trashItem `json:"files"`
}

type searchResult struct {
	Folders []folder `json:"folders"`
	Files   []file   `json:"files"`
}

// Whoami prints the current user.
func (c *Commands) Whoami() error {
	var u user
	if err := c.Client.Do(http.MethodGet, "/users/me", nil, &u); err != nil {
		return err
	}
	fmt.Fprintf(c.Out, "%s <%s>\n", u.DisplayName, u.Email)
	fmt.Fprintf(c.Out, "Storage: %s of %s\n", formatBytes(u.StorageUsed), formatBytes(u.StorageQuota))
	return nil
}

// Ls lists folders and files in a folder (current folder when folderID is empty).
func (c *Commands) Ls(folderID string) error {
	if folderID == "" {
		folderID = c.CurrentFolderID
	}
	var b browser
	if err := c.Client.Do(http.MethodGet, browserPath(folderID), nil, &b); err != nil {
		return err
	}
	for _, f := range b.Folders {
		fmt.Fprintf(c.Out, "/%s\n", f.Name)
	}
	for _, f := range b.Files {
		fmt.Fprintf(c.Out, "%s  %s\n", f.Name, formatBytes(f.SizeBytes))
	}
	return nil
}

// Upload uploads a local file, optionally into a folder (default: current folder).
func (c *Commands) Upload(localPath, folderID string) error {
	if folderID == "" {
		folderID = c.CurrentFolderID
	}
	info, err := os.Stat(localPath)
	if err != nil {
		return err
	}
	name := filepath.Base(localPath)
	contentType := mime.TypeByExtension(filepath.Ext(localPath))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	var session struct {
		FileID    string `json:"fileId"`
		UploadURL string `json:"uploadUrl"`
		ExpiresAt string `json:"expiresAt"`
	}
	req := map[string]any{
		"name":        name,
		"size":        info.Size(),
		"contentType": contentType,
	}
	if folderID != "" {
		req["folderId"] = folderID
	}
	if err := c.Client.Do(http.MethodPost, "/files/upload-sessions", req, &session); err != nil {
		return err
	}

	data, err := os.ReadFile(localPath)
	if err != nil {
		return err
	}
	if err := putBytes(session.UploadURL, contentType, data); err != nil {
		return err
	}

	var completed file
	if err := c.Client.Do(http.MethodPost, "/files/"+session.FileID+"/complete", nil, &completed); err != nil {
		return err
	}
	fmt.Fprintf(c.Out, "Uploaded %s\n", name)
	return nil
}

// UploadReplace uploads a local file, replacing an existing file by name.
func (c *Commands) UploadReplace(localPath, name string) error {
	info, err := os.Stat(localPath)
	if err != nil {
		return err
	}
	_, fileID, err := c.resolve(name)
	if err != nil {
		return err
	}
	contentType := mime.TypeByExtension(filepath.Ext(localPath))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	var session struct {
		FileID    string `json:"fileId"`
		UploadURL string `json:"uploadUrl"`
		ExpiresAt string `json:"expiresAt"`
	}
	req := map[string]any{
		"name":          name,
		"size":          info.Size(),
		"contentType":   contentType,
		"replaceFileId": fileID,
	}
	if err := c.Client.Do(http.MethodPost, "/files/upload-sessions", req, &session); err != nil {
		return err
	}

	data, err := os.ReadFile(localPath)
	if err != nil {
		return err
	}
	if err := putBytes(session.UploadURL, contentType, data); err != nil {
		return err
	}

	var completed file
	if err := c.Client.Do(http.MethodPost, "/files/"+session.FileID+"/complete", nil, &completed); err != nil {
		return err
	}
	fmt.Fprintf(c.Out, "Replaced %s\n", name)
	return nil
}

// Versions lists archived versions of a file by name.
func (c *Commands) Versions(name string) error {
	_, fileID, err := c.resolve(name)
	if err != nil {
		return err
	}
	var out struct {
		Versions []struct {
			ID        string `json:"id"`
			SizeBytes int64  `json:"sizeBytes"`
			MimeType  string `json:"mimeType"`
			CreatedAt string `json:"createdAt"`
		} `json:"versions"`
	}
	if err := c.Client.Do(http.MethodGet, "/files/"+fileID+"/versions", nil, &out); err != nil {
		return err
	}
	if len(out.Versions) == 0 {
		fmt.Fprintln(c.Out, "No versions")
		return nil
	}
	for _, v := range out.Versions {
		fmt.Fprintf(c.Out, "%s  %s  %s\n", v.ID, formatBytes(v.SizeBytes), v.CreatedAt)
	}
	return nil
}

// VersionDownload downloads an archived version of a file by name.
func (c *Commands) VersionDownload(name, versionID string) error {
	_, fileID, err := c.resolve(name)
	if err != nil {
		return err
	}
	var dl struct {
		DownloadURL string `json:"downloadUrl"`
		ExpiresAt   string `json:"expiresAt"`
	}
	if err := c.Client.Do(http.MethodGet, "/files/"+fileID+"/versions/"+versionID+"/download", nil, &dl); err != nil {
		return err
	}
	data, err := getBytes(dl.DownloadURL)
	if err != nil {
		return err
	}
	outName := name + "." + versionID
	if err := os.WriteFile(outName, data, 0o600); err != nil {
		return err
	}
	fmt.Fprintf(c.Out, "Downloaded %s\n", outName)
	return nil
}

// Logout revokes the current refresh token and clears saved tokens.
func (c *Commands) Logout(refreshToken string, clear func() error) error {
	if err := c.Client.Do(http.MethodPost, "/auth/logout", map[string]string{
		"refreshToken": refreshToken,
	}, nil); err != nil {
		return err
	}
	if clear != nil {
		if err := clear(); err != nil {
			return err
		}
	}
	fmt.Fprintln(c.Out, "Logged out")
	return nil
}

// Download downloads a file by name from the current folder.
func (c *Commands) Download(name string) error {
	var b browser
	if err := c.Client.Do(http.MethodGet, browserPath(c.CurrentFolderID), nil, &b); err != nil {
		return err
	}
	var target *file
	for i := range b.Files {
		if b.Files[i].Name == name {
			target = &b.Files[i]
			break
		}
	}
	if target == nil {
		return &client.Error{Code: "NOT_FOUND", Message: "file not found: " + name, Status: 404}
	}

	var dl struct {
		DownloadURL string `json:"downloadUrl"`
		ExpiresAt   string `json:"expiresAt"`
	}
	if err := c.Client.Do(http.MethodGet, "/files/"+target.ID+"/download", nil, &dl); err != nil {
		return err
	}

	data, err := getBytes(dl.DownloadURL)
	if err != nil {
		return err
	}
	if err := os.WriteFile(name, data, 0o600); err != nil {
		return err
	}
	fmt.Fprintf(c.Out, "Downloaded %s\n", name)
	return nil
}

func putBytes(url, contentType string, data []byte) error {
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", contentType)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		return &client.Error{Code: "UPLOAD_FAILED", Message: fmt.Sprintf("upload failed (%d)", res.StatusCode), Status: res.StatusCode}
	}
	return nil
}

func getBytes(url string) ([]byte, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		return nil, &client.Error{Code: "DOWNLOAD_FAILED", Message: fmt.Sprintf("download failed (%d)", res.StatusCode), Status: res.StatusCode}
	}
	return io.ReadAll(res.Body)
}

func formatBytes(n int64) string {
	switch {
	case n < 1024:
		return fmt.Sprintf("%d B", n)
	case n < 1024*1024:
		return fmt.Sprintf("%.1f KiB", float64(n)/1024)
	case n < 1024*1024*1024:
		return fmt.Sprintf("%.1f MiB", float64(n)/(1024*1024))
	default:
		return fmt.Sprintf("%.2f GiB", float64(n)/(1024*1024*1024))
	}
}

// browserPath builds the /browser path for a folder (root when empty).
func browserPath(folderID string) string {
	if folderID == "" {
		return "/browser"
	}
	return "/browser?folderId=" + folderID
}

// resolve finds a folder or file by exact name in the current folder.
// It returns the resource type ("folder" or "file") and its ID.
func (c *Commands) resolve(name string) (string, string, error) {
	var b browser
	if err := c.Client.Do(http.MethodGet, browserPath(c.CurrentFolderID), nil, &b); err != nil {
		return "", "", err
	}
	for _, f := range b.Folders {
		if f.Name == name {
			return "folder", f.ID, nil
		}
	}
	for _, f := range b.Files {
		if f.Name == name {
			return "file", f.ID, nil
		}
	}
	return "", "", &client.Error{Code: "NOT_FOUND", Message: "not found: " + name, Status: 404}
}

// Cd changes the current folder. Empty name returns to root; ".." goes up.
func (c *Commands) Cd(name string) error {
	if name == "" {
		if c.SaveFolder != nil {
			if err := c.SaveFolder(""); err != nil {
				return err
			}
		}
		c.CurrentFolderID = ""
		fmt.Fprintln(c.Out, "Now in /")
		return nil
	}
	var b browser
	if err := c.Client.Do(http.MethodGet, browserPath(c.CurrentFolderID), nil, &b); err != nil {
		return err
	}
	if name == ".." {
		parent := ""
		if b.Folder != nil && b.Folder.ParentID != nil {
			parent = *b.Folder.ParentID
		}
		if c.SaveFolder != nil {
			if err := c.SaveFolder(parent); err != nil {
				return err
			}
		}
		c.CurrentFolderID = parent
		if parent == "" {
			fmt.Fprintln(c.Out, "Now in /")
		} else {
			fmt.Fprintf(c.Out, "Now in %s\n", parent)
		}
		return nil
	}
	for _, f := range b.Folders {
		if f.Name == name {
			if c.SaveFolder != nil {
				if err := c.SaveFolder(f.ID); err != nil {
					return err
				}
			}
			c.CurrentFolderID = f.ID
			fmt.Fprintf(c.Out, "Now in %s\n", name)
			return nil
		}
	}
	return &client.Error{Code: "NOT_FOUND", Message: "folder not found: " + name, Status: 404}
}

// Pwd prints the current folder id (or "/" for root).
func (c *Commands) Pwd() error {
	if c.CurrentFolderID == "" {
		fmt.Fprintln(c.Out, "/")
	} else {
		fmt.Fprintln(c.Out, c.CurrentFolderID)
	}
	return nil
}

// Mkdir creates a folder, optionally under a parent (default: current folder).
func (c *Commands) Mkdir(name, parentID string) error {
	if parentID == "" {
		parentID = c.CurrentFolderID
	}
	req := map[string]any{"name": name}
	if parentID != "" {
		req["parentId"] = parentID
	}
	if err := c.Client.Do(http.MethodPost, "/folders", req, nil); err != nil {
		return err
	}
	fmt.Fprintf(c.Out, "Created folder %s\n", name)
	return nil
}

// Rm soft-deletes a file or folder by name from the root.
func (c *Commands) Rm(name string) error {
	typ, id, err := c.resolve(name)
	if err != nil {
		return err
	}
	path := "/files/" + id
	if typ == "folder" {
		path = "/folders/" + id
	}
	if err := c.Client.Do(http.MethodDelete, path, nil, nil); err != nil {
		return err
	}
	fmt.Fprintf(c.Out, "Deleted %s\n", name)
	return nil
}

// Mv renames (newName) or moves (toFolderID) a file/folder by name.
func (c *Commands) Mv(name, newName, toFolderID string) error {
	typ, id, err := c.resolve(name)
	if err != nil {
		return err
	}
	path := "/files/" + id
	if typ == "folder" {
		path = "/folders/" + id
	}
	var body map[string]any
	if newName != "" {
		body = map[string]any{"name": newName}
	} else {
		key := "folderId"
		if typ == "folder" {
			key = "parentId"
		}
		body = map[string]any{key: toFolderID}
	}
	if err := c.Client.Do(http.MethodPatch, path, body, nil); err != nil {
		return err
	}
	fmt.Fprintf(c.Out, "Moved %s\n", name)
	return nil
}

// Trash lists trashed folders and files.
func (c *Commands) Trash() error {
	var t trashList
	if err := c.Client.Do(http.MethodGet, "/trash", nil, &t); err != nil {
		return err
	}
	if len(t.Folders) == 0 && len(t.Files) == 0 {
		fmt.Fprintln(c.Out, "(empty)")
		return nil
	}
	for _, f := range t.Folders {
		fmt.Fprintf(c.Out, "/%s\n", f.Name)
	}
	for _, f := range t.Files {
		size := int64(0)
		if f.SizeBytes != nil {
			size = *f.SizeBytes
		}
		fmt.Fprintf(c.Out, "%s  %s\n", f.Name, formatBytes(size))
	}
	return nil
}

// Restore restores a trashed file or folder by name.
func (c *Commands) Restore(name string) error {
	var t trashList
	if err := c.Client.Do(http.MethodGet, "/trash", nil, &t); err != nil {
		return err
	}
	for _, f := range t.Folders {
		if f.Name == name {
			if err := c.Client.Do(http.MethodPost, "/folders/"+f.ID+"/restore", nil, nil); err != nil {
				return err
			}
			fmt.Fprintf(c.Out, "Restored %s\n", name)
			return nil
		}
	}
	for _, f := range t.Files {
		if f.Name == name {
			if err := c.Client.Do(http.MethodPost, "/files/"+f.ID+"/restore", nil, nil); err != nil {
				return err
			}
			fmt.Fprintf(c.Out, "Restored %s\n", name)
			return nil
		}
	}
	return &client.Error{Code: "NOT_FOUND", Message: "not found in trash: " + name, Status: 404}
}

// Search lists folders and files matching a query.
func (c *Commands) Search(q string) error {
	var r searchResult
	if err := c.Client.Do(http.MethodGet, "/search?q="+q, nil, &r); err != nil {
		return err
	}
	for _, f := range r.Folders {
		fmt.Fprintf(c.Out, "/%s\n", f.Name)
	}
	for _, f := range r.Files {
		fmt.Fprintf(c.Out, "%s  %s\n", f.Name, formatBytes(f.SizeBytes))
	}
	return nil
}

// Storage prints used vs quota bytes.
func (c *Commands) Storage() error {
	var s struct {
		UsedBytes  int64 `json:"usedBytes"`
		QuotaBytes int64 `json:"quotaBytes"`
	}
	if err := c.Client.Do(http.MethodGet, "/storage", nil, &s); err != nil {
		return err
	}
	fmt.Fprintf(c.Out, "%s of %s\n", formatBytes(s.UsedBytes), formatBytes(s.QuotaBytes))
	return nil
}

// Purge permanently deletes a trashed file or folder by name.
func (c *Commands) Purge(name string) error {
	var t trashList
	if err := c.Client.Do(http.MethodGet, "/trash", nil, &t); err != nil {
		return err
	}
	for _, f := range t.Folders {
		if f.Name == name {
			if err := c.Client.Do(http.MethodDelete, "/trash/folders/"+f.ID, nil, nil); err != nil {
				return err
			}
			fmt.Fprintf(c.Out, "Purged %s\n", name)
			return nil
		}
	}
	for _, f := range t.Files {
		if f.Name == name {
			if err := c.Client.Do(http.MethodDelete, "/trash/files/"+f.ID, nil, nil); err != nil {
				return err
			}
			fmt.Fprintf(c.Out, "Purged %s\n", name)
			return nil
		}
	}
	return &client.Error{Code: "NOT_FOUND", Message: "not found in trash: " + name, Status: 404}
}

// LsGlob lists files in the current folder matching a glob pattern.
func (c *Commands) LsGlob(pattern string) error {
	var b browser
	if err := c.Client.Do(http.MethodGet, browserPath(c.CurrentFolderID), nil, &b); err != nil {
		return err
	}
	for _, f := range b.Files {
		ok, err := filepath.Match(pattern, f.Name)
		if err != nil {
			return err
		}
		if ok {
			fmt.Fprintf(c.Out, "%s  %s\n", f.Name, formatBytes(f.SizeBytes))
		}
	}
	return nil
}

type shareOutgoing struct {
	ID           string `json:"id"`
	ResourceType string `json:"resourceType"`
	ResourceID   string `json:"resourceId"`
	ResourceName string `json:"resourceName"`
	Recipient    struct {
		Email string `json:"email"`
	} `json:"recipient"`
}

type shareIncoming struct {
	ID           string `json:"id"`
	ResourceType string `json:"resourceType"`
	ResourceID   string `json:"resourceId"`
	ResourceName string `json:"resourceName"`
	Owner        struct {
		Email string `json:"email"`
	} `json:"owner"`
}

// Share shares a file or folder (by name in the current folder) with an email.
func (c *Commands) Share(name, email string) error {
	typ, id, err := c.resolve(name)
	if err != nil {
		return err
	}
	if err := c.Client.Do(http.MethodPost, "/shares", map[string]string{
		"resourceType": typ,
		"resourceId":   id,
		"email":        email,
	}, nil); err != nil {
		return err
	}
	fmt.Fprintf(c.Out, "Shared %s with %s\n", name, email)
	return nil
}

// Shares lists shares the current user created.
func (c *Commands) Shares() error {
	var out struct {
		Shares []shareOutgoing `json:"shares"`
	}
	if err := c.Client.Do(http.MethodGet, "/shares", nil, &out); err != nil {
		return err
	}
	if len(out.Shares) == 0 {
		fmt.Fprintln(c.Out, "(empty)")
		return nil
	}
	for _, s := range out.Shares {
		fmt.Fprintf(c.Out, "%s  %s  -> %s\n", s.ID, s.ResourceName, s.Recipient.Email)
	}
	return nil
}

// SharedWithMe lists shares the current user received.
func (c *Commands) SharedWithMe() error {
	var out struct {
		Shares []shareIncoming `json:"shares"`
	}
	if err := c.Client.Do(http.MethodGet, "/shares/with-me", nil, &out); err != nil {
		return err
	}
	if len(out.Shares) == 0 {
		fmt.Fprintln(c.Out, "(empty)")
		return nil
	}
	for _, s := range out.Shares {
		fmt.Fprintf(c.Out, "%s  %s  <- %s\n", s.ID, s.ResourceName, s.Owner.Email)
	}
	return nil
}

// Unshare revokes a share by id.
func (c *Commands) Unshare(id string) error {
	if err := c.Client.Do(http.MethodDelete, "/shares/"+id, nil, nil); err != nil {
		return err
	}
	fmt.Fprintf(c.Out, "Revoked share %s\n", id)
	return nil
}

// SharedLs browses a folder shared with the current user.
func (c *Commands) SharedLs(folderID string) error {
	var b struct {
		Folder  *folder  `json:"folder"`
		Folders []folder `json:"folders"`
		Files   []file   `json:"files"`
	}
	if err := c.Client.Do(http.MethodGet, "/shared/folders/"+folderID, nil, &b); err != nil {
		return err
	}
	for _, f := range b.Folders {
		fmt.Fprintf(c.Out, "/%s\n", f.Name)
	}
	for _, f := range b.Files {
		fmt.Fprintf(c.Out, "%s  %s\n", f.Name, formatBytes(f.SizeBytes))
	}
	return nil
}

// SharedDownload downloads a file shared with the current user.
func (c *Commands) SharedDownload(fileID, name string) error {
	var dl struct {
		DownloadURL string `json:"downloadUrl"`
		ExpiresAt   string `json:"expiresAt"`
	}
	if err := c.Client.Do(http.MethodGet, "/shared/files/"+fileID+"/download", nil, &dl); err != nil {
		return err
	}
	data, err := getBytes(dl.DownloadURL)
	if err != nil {
		return err
	}
	if err := os.WriteFile(name, data, 0o600); err != nil {
		return err
	}
	fmt.Fprintf(c.Out, "Downloaded %s\n", name)
	return nil
}
