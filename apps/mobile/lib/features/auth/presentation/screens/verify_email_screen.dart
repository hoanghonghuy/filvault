import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../../../core/theme/app_colors.dart';
import '../../../../core/widgets/app_button.dart';
import '../../../../core/widgets/app_text_field.dart';
import '../controllers/auth_controller.dart';

class VerifyEmailScreen extends ConsumerStatefulWidget {
  const VerifyEmailScreen({super.key});

  @override
  ConsumerState<VerifyEmailScreen> createState() => _VerifyEmailScreenState();
}

class _VerifyEmailScreenState extends ConsumerState<VerifyEmailScreen> {
  final _codeController = TextEditingController();
  bool _resending = false;
  String? _resendSuccessMessage;

  @override
  void dispose() {
    _codeController.dispose();
    super.dispose();
  }

  Future<void> _handleVerify() async {
    final code = _codeController.text.trim();
    if (code.isEmpty) return;
    FocusScope.of(context).unfocus();

    final success = await ref.read(authControllerProvider.notifier).verifyEmail(code);
    if (success && mounted) {
      context.go('/files');
    }
  }

  Future<void> _handleResend() async {
    setState(() {
      _resending = true;
      _resendSuccessMessage = null;
    });
    await ref.read(authControllerProvider.notifier).resendVerification();
    if (mounted) {
      setState(() {
        _resending = false;
        _resendSuccessMessage = 'Đã gửi lại mã xác thực vào email của bạn.';
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    final authState = ref.watch(authControllerProvider);
    final userEmail = authState.user?.email ?? '';

    return Scaffold(
      backgroundColor: AppColors.canvas,
      appBar: AppBar(
        title: const Text('Xác thực Email'),
        actions: [
          TextButton(
            onPressed: () async {
              await ref.read(authControllerProvider.notifier).logout();
              if (context.mounted) context.go('/login');
            },
            child: const Text('Đăng xuất', style: TextStyle(color: AppColors.muted)),
          ),
        ],
      ),
      body: SafeArea(
        child: SingleChildScrollView(
          padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 20),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              const SizedBox(height: 20),
              const Center(
                child: Icon(
                  Icons.mark_email_unread_outlined,
                  size: 64,
                  color: AppColors.accent,
                ),
              ),
              const SizedBox(height: 20),
              const Text(
                'Kiểm tra hòm thư của bạn',
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 22,
                  fontWeight: FontWeight.w700,
                  color: AppColors.ink,
                ),
              ),
              const SizedBox(height: 8),
              Text(
                'Chúng tôi đã gửi mã xác thực đến địa chỉ:\n$userEmail',
                textAlign: TextAlign.center,
                style: const TextStyle(fontSize: 14, color: AppColors.muted, height: 1.4),
              ),
              const SizedBox(height: 32),

              if (_resendSuccessMessage != null) ...[
                Container(
                  padding: const EdgeInsets.all(12),
                  decoration: BoxDecoration(
                    color: AppColors.accentSoft,
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: Text(
                    _resendSuccessMessage!,
                    textAlign: TextAlign.center,
                    style: const TextStyle(color: AppColors.accentHover, fontSize: 13, fontWeight: FontWeight.w500),
                  ),
                ),
                const SizedBox(height: 20),
              ],

              AppTextField(
                controller: _codeController,
                label: 'Mã xác thực',
                hint: 'Nhập mã gồm các chữ số',
                keyboardType: TextInputType.text,
                prefixIcon: Icons.security_outlined,
                textInputAction: TextInputAction.done,
                onFieldSubmitted: (_) => _handleVerify(),
              ),
              const SizedBox(height: 24),

              AppButton(
                label: 'Xác nhận mã',
                isLoading: authState.isLoading,
                onPressed: _handleVerify,
              ),
              const SizedBox(height: 16),

              TextButton(
                onPressed: _resending ? null : _handleResend,
                child: _resending
                    ? const SizedBox(
                        width: 16,
                        height: 16,
                        child: CircularProgressIndicator(strokeWidth: 2),
                      )
                    : const Text(
                        'Chưa nhận được mã? Gửi lại',
                        style: TextStyle(
                          fontSize: 14,
                          fontWeight: FontWeight.w600,
                          color: AppColors.accent,
                        ),
                      ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
