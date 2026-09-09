import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:local_auth/local_auth.dart';
import '../../../../core/theme/app_colors.dart';
import '../../../activity/presentation/screens/activity_screen.dart';
import '../../../auth/presentation/controllers/auth_controller.dart';
import '../../../files/presentation/widgets/file_icon_helper.dart';
import '../../../shares/presentation/screens/shares_screen.dart';
import '../../../trash/presentation/screens/trash_screen.dart';

class SettingsScreen extends ConsumerStatefulWidget {
  const SettingsScreen({super.key});

  @override
  ConsumerState<SettingsScreen> createState() => _SettingsScreenState();
}

class _SettingsScreenState extends ConsumerState<SettingsScreen> {
  final LocalAuthentication _localAuth = LocalAuthentication();
  bool _biometricsEnabled = false;

  @override
  Widget build(BuildContext context) {
    final authState = ref.watch(authControllerProvider);
    final user = authState.user;
    final theme = Theme.of(context);

    final usedStr = FileIconHelper.formatBytes(user?.storageUsed ?? 0);
    final quotaStr = FileIconHelper.formatBytes(user?.storageQuota ?? (10 * 1024 * 1024 * 1024));

    return Scaffold(
      backgroundColor: theme.scaffoldBackgroundColor,
      appBar: AppBar(
        title: const Text('Cài đặt', style: TextStyle(fontWeight: FontWeight.w700)),
      ),
      body: ListView(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
        children: [
          // User Profile Card
          Container(
            padding: const EdgeInsets.all(16),
            decoration: BoxDecoration(
              color: theme.cardColor,
              borderRadius: BorderRadius.circular(16),
              boxShadow: [
                BoxShadow(
                  color: Colors.black.withOpacity(0.04),
                  blurRadius: 10,
                  offset: const Offset(0, 2),
                ),
              ],
            ),
            child: Row(
              children: [
                CircleAvatar(
                  radius: 30,
                  backgroundColor: AppColors.accent,
                  backgroundImage: user?.avatarUrl != null ? NetworkImage(user!.avatarUrl!) : null,
                  child: user?.avatarUrl == null
                      ? Text(
                          (user?.displayName.isNotEmpty == true ? user!.displayName : user?.email ?? 'U')[0].toUpperCase(),
                          style: const TextStyle(fontSize: 22, fontWeight: FontWeight.w700, color: Colors.white),
                        )
                      : null,
                ),
                const SizedBox(width: 14),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        user?.displayName.isNotEmpty == true ? user!.displayName : 'Người dùng Filvault',
                        style: const TextStyle(fontSize: 16, fontWeight: FontWeight.w700),
                      ),
                      const SizedBox(height: 2),
                      Text(
                        user?.email ?? '',
                        style: TextStyle(fontSize: 13, color: theme.colorScheme.outline),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(height: 16),

          // Storage Quota Card
          Container(
            padding: const EdgeInsets.all(16),
            decoration: BoxDecoration(
              color: theme.cardColor,
              borderRadius: BorderRadius.circular(16),
              boxShadow: [
                BoxShadow(
                  color: Colors.black.withOpacity(0.04),
                  blurRadius: 10,
                  offset: const Offset(0, 2),
                ),
              ],
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    const Text('Dung lượng lưu trữ', style: TextStyle(fontSize: 14, fontWeight: FontWeight.w600)),
                    Text('$usedStr / $quotaStr', style: TextStyle(fontSize: 13, color: theme.colorScheme.outline)),
                  ],
                ),
                const SizedBox(height: 10),
                ClipRRect(
                  borderRadius: BorderRadius.circular(4),
                  child: LinearProgressIndicator(
                    value: user != null && user.storageQuota > 0
                        ? (user.storageUsed / user.storageQuota).clamp(0.0, 1.0)
                        : 0.0,
                    minHeight: 8,
                    backgroundColor: theme.dividerColor.withOpacity(0.1),
                    valueColor: const AlwaysStoppedAnimation<Color>(AppColors.accent),
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(height: 24),

          // Security & Preferences Header
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 4, vertical: 8),
            child: Text(
              'Bảo mật & Ứng dụng',
              style: TextStyle(
                fontWeight: FontWeight.bold,
                fontSize: 13,
                color: theme.colorScheme.primary,
              ),
            ),
          ),

          // Biometric Lock Toggle
          SwitchListTile(
            secondary: const Icon(Icons.fingerprint_rounded),
            title: const Text('Khóa ứng dụng bằng sinh trắc học'),
            subtitle: const Text('Vân tay hoặc khuôn mặt khi mở ứng dụng'),
            value: _biometricsEnabled,
            activeColor: AppColors.accent,
            onChanged: (val) async {
              final canCheck = await _localAuth.canCheckBiometrics;
              if (canCheck) {
                final authenticated = await _localAuth.authenticate(
                  localizedReason: 'Xác thực để bật khóa sinh trắc học',
                );
                if (authenticated) {
                  setState(() => _biometricsEnabled = val);
                }
              } else {
                if (mounted) {
                  ScaffoldMessenger.of(context).showSnackBar(
                    const SnackBar(content: Text('Thiết bị chưa cài đặt vân tay/khuôn mặt')),
                  );
                }
              }
            },
          ),

          // Features Header
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 4, vertical: 8),
            child: Text(
              'Quản lý nội dung',
              style: TextStyle(
                fontWeight: FontWeight.bold,
                fontSize: 13,
                color: theme.colorScheme.primary,
              ),
            ),
          ),

          // Trash
          ListTile(
            leading: const Icon(Icons.delete_outline_rounded),
            title: const Text('Thùng rác'),
            subtitle: const Text('Xem và khôi phục các mục đã xóa'),
            trailing: const Icon(Icons.chevron_right_rounded),
            onTap: () {
              Navigator.push(
                context,
                MaterialPageRoute(builder: (_) => const TrashScreen()),
              );
            },
          ),

          // Activity
          ListTile(
            leading: const Icon(Icons.history_rounded),
            title: const Text('Nhật ký hoạt động'),
            subtitle: const Text('Lịch sử tải lên, xóa và thao tác'),
            trailing: const Icon(Icons.chevron_right_rounded),
            onTap: () {
              Navigator.push(
                context,
                MaterialPageRoute(builder: (_) => const ActivityScreen()),
              );
            },
          ),

          // Shares
          ListTile(
            leading: const Icon(Icons.share_rounded),
            title: const Text('Quản lý chia sẻ'),
            subtitle: const Text('Mục chia sẻ nội bộ đến bạn hoặc từ bạn'),
            trailing: const Icon(Icons.chevron_right_rounded),
            onTap: () {
              Navigator.push(
                context,
                MaterialPageRoute(builder: (_) => const SharesScreen()),
              );
            },
          ),

          const Divider(height: 24),

          // Change Password
          ListTile(
            leading: const Icon(Icons.lock_reset_rounded),
            title: const Text('Đổi mật khẩu'),
            trailing: const Icon(Icons.chevron_right_rounded),
            onTap: () => _showChangePasswordDialog(context),
          ),

          // Clear cache
          ListTile(
            leading: const Icon(Icons.cleaning_services_rounded),
            title: const Text('Dọn sạch bộ nhớ đệm (Cache)'),
            subtitle: const Text('Xóa các hình ảnh thumbnail đã lưu tạm'),
            trailing: const Icon(Icons.chevron_right_rounded),
            onTap: () {
              PaintingBinding.instance.imageCache.clear();
              PaintingBinding.instance.imageCache.clearLiveImages();
              ScaffoldMessenger.of(context).showSnackBar(
                const SnackBar(content: Text('Đã làm trống bộ nhớ đệm ứng dụng')),
              );
            },
          ),

          const Divider(height: 32),

          // Logout
          ListTile(
            leading: Icon(Icons.logout_rounded, color: theme.colorScheme.error),
            title: Text(
              'Đăng xuất',
              style: TextStyle(color: theme.colorScheme.error, fontWeight: FontWeight.w600),
            ),
            shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
            onTap: () async {
              final confirm = await showDialog<bool>(
                context: context,
                builder: (ctx) => AlertDialog(
                  title: const Text('Đăng xuất'),
                  content: const Text('Bạn có chắc chắn muốn đăng xuất khỏi tài khoản này?'),
                  actions: [
                    TextButton(
                      onPressed: () => Navigator.of(ctx).pop(false),
                      child: const Text('Hủy'),
                    ),
                    FilledButton(
                      style: FilledButton.styleFrom(backgroundColor: theme.colorScheme.error),
                      onPressed: () => Navigator.of(ctx).pop(true),
                      child: const Text('Đăng xuất'),
                    ),
                  ],
                ),
              );
              if (confirm == true && context.mounted) {
                await ref.read(authControllerProvider.notifier).logout();
                if (context.mounted) context.go('/login');
              }
            },
          ),
        ],
      ),
    );
  }

  void _showChangePasswordDialog(BuildContext context) {
    final oldPassController = TextEditingController();
    final newPassController = TextEditingController();

    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Đổi mật khẩu'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            TextField(
              controller: oldPassController,
              obscureText: true,
              decoration: const InputDecoration(labelText: 'Mật khẩu hiện tại'),
            ),
            const SizedBox(height: 12),
            TextField(
              controller: newPassController,
              obscureText: true,
              decoration: const InputDecoration(labelText: 'Mật khẩu mới (tối thiểu 8 ký tự)'),
            ),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx),
            child: const Text('Hủy'),
          ),
          FilledButton(
            onPressed: () async {
              final oldPass = oldPassController.text;
              final newPass = newPassController.text;
              if (newPass.length < 8) {
                ScaffoldMessenger.of(context).showSnackBar(
                  const SnackBar(content: Text('Mật khẩu mới phải từ 8 ký tự')),
                );
                return;
              }
              Navigator.pop(ctx);
              final success = await ref
                  .read(authControllerProvider.notifier)
                  .changePassword(oldPass, newPass);
              if (context.mounted) {
                ScaffoldMessenger.of(context).showSnackBar(
                  SnackBar(
                    content: Text(success ? 'Đổi mật khẩu thành công' : 'Đổi mật khẩu thất bại'),
                  ),
                );
              }
            },
            child: const Text('Lưu'),
          ),
        ],
      ),
    );
  }
}
