import 'dart:async';
import 'dart:convert';
import 'package:dio/dio.dart';
import '../../../../core/api/api_client.dart';
import '../../../../core/api/api_endpoints.dart';

class ChatSseEvent {
  final int? sequence;
  final String type;
  final Map<String, dynamic> data;

  ChatSseEvent({
    this.sequence,
    required this.type,
    required this.data,
  });
}

class ChatSseService {
  final ApiClient _apiClient;
  StreamController<ChatSseEvent>? _controller;
  CancelToken? _cancelToken;
  bool _isRunning = false;
  int _lastSequence = 0;

  ChatSseService({ApiClient? apiClient})
      : _apiClient = apiClient ?? ApiClient.instance;

  Stream<ChatSseEvent> get eventStream {
    _controller ??= StreamController<ChatSseEvent>.broadcast(
      onListen: _startListening,
      onCancel: _stopListening,
    );
    return _controller!.stream;
  }

  void _startListening() {
    if (_isRunning) return;
    _isRunning = true;
    _connect();
  }

  void _stopListening() {
    _isRunning = false;
    _cancelToken?.cancel();
    _cancelToken = null;
  }

  Future<void> _connect() async {
    while (_isRunning) {
      try {
        _cancelToken = CancelToken();
        final headers = <String, dynamic>{
          'Accept': 'text/event-stream',
          'Cache-Control': 'no-cache',
        };
        if (_lastSequence > 0) {
          headers['Last-Event-ID'] = '$_lastSequence';
        }

        final response = await _apiClient.dio.get<ResponseBody>(
          ApiEndpoints.chatEvents,
          queryParameters: _lastSequence > 0 ? {'after': _lastSequence} : null,
          options: Options(
            responseType: ResponseType.stream,
            headers: headers,
          ),
          cancelToken: _cancelToken,
        );

        String currentEvent = 'message';
        int? currentId;
        final buffer = StringBuffer();

        final stream = response.data!.stream
            .cast<List<int>>()
            .transform(utf8.decoder)
            .transform(const LineSplitter());

        await for (final line in stream) {
          if (!_isRunning) break;

          if (line.isEmpty) {
            // End of event block
            if (buffer.isNotEmpty) {
              try {
                final jsonMap = jsonDecode(buffer.toString()) as Map<String, dynamic>;
                _controller?.add(ChatSseEvent(
                  sequence: currentId,
                  type: currentEvent,
                  data: jsonMap,
                ));
              } catch (_) {}
              buffer.clear();
            }
            currentEvent = 'message';
            continue;
          }

          if (line.startsWith('id:')) {
            final idStr = line.substring(3).trim();
            final parsed = int.tryParse(idStr);
            if (parsed != null) {
              currentId = parsed;
              _lastSequence = parsed;
            }
          } else if (line.startsWith('event:')) {
            currentEvent = line.substring(6).trim();
          } else if (line.startsWith('data:')) {
            buffer.write(line.substring(5).trim());
          }
        }
      } catch (e) {
        if (!_isRunning) break;
        // Wait 3 seconds before reconnecting
        await Future.delayed(const Duration(seconds: 3));
      }
    }
  }

  void dispose() {
    _stopListening();
    _controller?.close();
    _controller = null;
  }
}
