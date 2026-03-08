/**
 * Voice Chat Module - Client-side voice support for Learning Desktop
 *
 * Features:
 * - Speech-to-Text using Web Speech API (SpeechRecognition)
 * - Text-to-Speech using Web Speech API (SpeechSynthesis)
 * - Voice Activity Detection (VAD) using AudioWorklet
 * - Student speech analysis for learning insights
 *
 * Browser Support:
 * - Chrome/Edge: Full support
 * - Safari: Partial (no SpeechRecognition)
 * - Firefox: No SpeechRecognition API
 *
 * Fallback: Server-side processing available for unsupported browsers
 */

(function(window) {
    'use strict';

    // Voice Chat Constructor
    function VoiceChat(options) {
        this.options = Object.assign({
            inputSelector: '#chat-input',
            micButtonSelector: '[data-voice-mic]',
            speakerSelector: '[data-voice-speaker]',
            onTranscript: null,
            onSpeakingStart: null,
            onSpeakingEnd: null,
            onError: null,
            debug: false
        }, options);

        this.isListening = false;
        this.isSpeaking = false;
        this.recognition = null;
        this.synthesis = window.speechSynthesis;
        this.vadAnalyzer = null;

        this.init();
    }

    VoiceChat.prototype = {
        init: function() {
            this.log('Initializing VoiceChat...');

            // Check browser support
            const SpeechRecognition = window.SpeechRecognition || window.webkitSpeechRecognition;
            const speechRecognitionSupported = !!SpeechRecognition;
            const speechSynthesisSupported = !!this.synthesis;

            this.log('SpeechRecognition:', speechRecognitionSupported);
            this.log('SpeechSynthesis:', speechSynthesisSupported);

            if (!speechRecognitionSupported && !speechSynthesisSupported) {
                this.log('No voice support available, falling back to server');
                this.fallbackToServer();
                return;
            }

            // Setup Speech Recognition
            if (speechRecognitionSupported) {
                this.setupRecognition(SpeechRecognition);
            }

            // Setup Speech Synthesis
            if (speechSynthesisSupported) {
                this.setupSynthesis();
            }

            // Setup UI bindings
            this.setupUI();

            // Setup VAD if supported
            this.setupVAD();
        },

        setupRecognition: function(SpeechRecognition) {
            this.recognition = new SpeechRecognition();
            this.recognition.continuous = true;
            this.recognition.interimResults = true;
            this.recognition.lang = 'en-US';
            this.recognition.maxAlternatives = 1;

            var self = this;

            this.recognition.onstart = function() {
                self.isListening = true;
                self.log('Recognition started');
                self.updateMicButton(true);
                if (self.options.onSpeakingStart) {
                    self.options.onSpeakingStart();
                }
            };

            this.recognition.onend = function() {
                self.isListening = false;
                self.log('Recognition ended');
                self.updateMicButton(false);
                if (self.options.onSpeakingEnd) {
                    self.options.onSpeakingEnd();
                }
                // Auto-restart if we were manually stopped
                if (self.shouldRestart) {
                    self.startListening();
                }
            };

            this.recognition.onresult = function(event) {
                var interimTranscript = '';
                var finalTranscript = '';

                for (var i = event.resultIndex; i < event.results.length; i++) {
                    var transcript = event.results[i][0].transcript;
                    if (event.results[i].isFinal) {
                        finalTranscript += transcript;
                        self.log('Final transcript:', finalTranscript);
                        self.handleTranscript(finalTranscript);
                    } else {
                        interimTranscript += transcript;
                        self.log('Interim transcript:', interimTranscript);
                        self.showInterim(interimTranscript);
                    }
                }
            };

            this.recognition.onerror = function(event) {
                self.log('Recognition error:', event.error);
                if (event.error === 'no-speech') {
                    // No speech detected, restart
                    return;
                }
                if (self.options.onError) {
                    self.options.onError(event.error);
                }
            };
        },

        setupSynthesis: function() {
            // Preload voices
            if (this.synthesis.getVoices) {
                this.synthesis.getVoices();
                this.synthesis.onvoiceschanged = function() {
                    this.log('Voices loaded:', this.synthesis.getVoices().length);
                }.bind(this);
            }
        },

        setupVAD: function() {
            // Voice Activity Detection using AudioContext
            // This is a simplified VAD - for production, use WebRTC VAD or Silero
            if (!window.AudioContext && !window.webkitAudioContext) {
                this.log('Web Audio API not available, VAD disabled');
                return;
            }

            try {
                this.audioContext = new (window.AudioContext || window.webkitAudioContext)();
                this.analyzer = this.audioContext.createAnalyser();
                this.analyzer.fftSize = 2048;
                this.vadThreshold = 0.02; // Adjust based on testing
                this.log('VAD initialized');
            } catch (e) {
                this.log('VAD setup failed:', e);
            }
        },

        setupUI: function() {
            var self = this;

            // Mic button
            var micBtn = document.querySelector(this.options.micButtonSelector);
            if (micBtn) {
                micBtn.addEventListener('click', function(e) {
                    e.preventDefault();
                    self.toggleListening();
                });
            }

            // Speaker toggle
            var speakerBtn = document.querySelector(this.options.speakerSelector);
            if (speakerBtn) {
                speakerBtn.addEventListener('click', function(e) {
                    e.preventDefault();
                    self.toggleMute();
                });
            }
        },

        toggleListening: function() {
            if (this.isListening) {
                this.stopListening();
            } else {
                this.startListening();
            }
        },

        startListening: function() {
            this.log('Starting to listen...');
            this.shouldRestart = true;
            try {
                this.recognition.start();
            } catch (e) {
                this.log('Already listening');
            }
        },

        stopListening: function() {
            this.log('Stopping listening...');
            this.shouldRestart = false;
            try {
                this.recognition.stop();
            } catch (e) {
                this.log('Not listening');
            }
        },

        handleTranscript: function(transcript) {
            // Clean up transcript
            transcript = transcript.trim()
                .replace(/\s*\.\s*/g, '. ')  // Fix spacing around periods
                .replace(/\s*,\s*/g, ', ')  // Fix spacing around commas
                .replace(/\.+$/, '');      // Remove trailing periods

            if (!transcript) return;

            // Check for voice commands
            if (this.handleVoiceCommand(transcript)) {
                return;
            }

            // Insert into chat input
            var input = document.querySelector(this.options.inputSelector);
            if (input) {
                input.value = transcript;
                input.dispatchEvent(new Event('input', { bubbles: true }));

                // Optionally auto-submit
                // input.form.dispatchEvent(new Event('submit', { cancelable: true }));
            }

            // Callback
            if (this.options.onTranscript) {
                this.options.onTranscript(transcript);
            }
        },

        showInterim: function(transcript) {
            // Show interim transcript in UI
            var interimEl = document.querySelector('[data-voice-interim]');
            if (interimEl) {
                interimEl.textContent = transcript;
                interimEl.style.display = 'block';
            }
        },

        handleVoiceCommand: function(transcript) {
            var cmd = transcript.toLowerCase().trim();

            // Voice commands
            switch (cmd) {
                case 'stop listening':
                case 'stop':
                    this.stopListening();
                    return true;

                case 'start listening':
                case 'listen':
                    this.startListening();
                    return true;

                case 'submit':
                case 'send':
                    var input = document.querySelector(this.options.inputSelector);
                    if (input && input.form) {
                        input.form.dispatchEvent(new Event('submit', { cancelable: true }));
                    }
                    return true;

                case 'clear':
                    input.value = '';
                    return true;

                default:
                    return false;
            }
        },

        speak: function(text, options) {
            options = options || {};

            if (!this.synthesis) {
                this.log('Speech synthesis not available');
                return;
            }

            // Cancel any ongoing speech
            this.synthesis.cancel();

            var utterance = new SpeechSynthesisUtterance(text);

            // Voice selection - prefer "natural" voices
            var voices = this.synthesis.getVoices();
            var preferredVoices = [
                'Google US English',
                'Microsoft David',
                'Microsoft Zira',
                'Samantha',
                'Alex'
            ];

            for (var i = 0; i < preferredVoices.length; i++) {
                for (var j = 0; j < voices.length; j++) {
                    if (voices[j].name.includes(preferredVoices[i])) {
                        utterance.voice = voices[j];
                        break;
                    }
                }
                if (utterance.voice) break;
            }

            // Speech parameters
            utterance.rate = options.rate || 1.0;      // 0.1 to 10
            utterance.pitch = options.pitch || 1.0;    // 0 to 2
            utterance.volume = options.volume || 1.0;  // 0 to 1

            // Emotion simulation through rate/pitch
            if (options.emotion) {
                switch (options.emotion) {
                    case 'excited':
                        utterance.rate = 1.1;
                        utterance.pitch = 1.1;
                        break;
                    case 'calm':
                        utterance.rate = 0.9;
                        utterance.pitch = 0.95;
                        break;
                    case 'confused':
                        utterance.rate = 0.95;
                        utterance.pitch = 0.9;
                        break;
                }
            }

            var self = this;
            utterance.onstart = function() {
                self.isSpeaking = true;
                self.log('Speaking:', text);
                self.updateSpeakerButton(true);
            };

            utterance.onend = function() {
                self.isSpeaking = false;
                self.updateSpeakerButton(false);
            };

            utterance.onerror = function(e) {
                self.log('Speech error:', e);
                self.isSpeaking = false;
                self.updateSpeakerButton(false);
            };

            this.synthesis.speak(utterance);
        },

        stopSpeaking: function() {
            if (this.synthesis) {
                this.synthesis.cancel();
                this.isSpeaking = false;
                this.updateSpeakerButton(false);
            }
        },

        toggleMute: function() {
            this.isMuted = !this.isMuted;
            if (this.isMuted) {
                this.stopSpeaking();
            }
            this.updateSpeakerButton(!this.isMuted);
        },

        updateMicButton: function(isActive) {
            var btn = document.querySelector(this.options.micButtonSelector);
            if (btn) {
                if (isActive) {
                    btn.classList.add('listening');
                    btn.setAttribute('aria-label', 'Stop listening');
                    // Pulse animation
                    btn.style.animation = 'pulse 1.5s infinite';
                } else {
                    btn.classList.remove('listening');
                    btn.setAttribute('aria-label', 'Start listening');
                    btn.style.animation = '';
                }
            }
        },

        updateSpeakerButton: function(isActive) {
            var btn = document.querySelector(this.options.speakerSelector);
            if (btn) {
                if (isActive) {
                    btn.classList.add('speaking');
                } else {
                    btn.classList.remove('speaking');
                }
            }
        },

        // Speech Analysis for Learning
        analyzeSpeech: function(transcript, timingInfo) {
            var analysis = {
                transcript: transcript,
                timestamp: new Date().toISOString(),
                words: [],
                pace: {},
                fillers: [],
                sentiment: 'neutral',
                understandingScore: 0.5
            };

            // Word-level analysis
            var words = transcript.split(/\s+/);
            var wordCount = words.length;
            var duration = timingInfo.duration || 5000; // ms

            // Calculate pace (words per minute)
            analysis.pace.wordsPerMinute = Math.round((wordCount / duration) * 60000);
            if (analysis.pace.wordsPerMinute < 120) {
                analysis.pace.label = 'slow';
            } else if (analysis.pace.wordsPerMinute > 160) {
                analysis.pace.label = 'fast';
            } else {
                analysis.pace.label = 'normal';
            }

            // Detect filler words
            var fillerPatterns = ['um', 'uh', 'like', 'you know', 'actually', 'basically'];
            words.forEach(function(word, index) {
                var lowerWord = word.toLowerCase().trim();
                if (fillerPatterns.indexOf(lowerWord) !== -1) {
                    analysis.fillers.push({
                        word: word,
                        position: index
                    });
                }
            });

            // Sentiment analysis
            var confidentWords = ['understand', 'got it', 'clear', 'makes sense', 'yes', 'right'];
            var confusedWords = ['confused', "don't understand", 'lost', 'not sure', 'huh'];
            var questionWords = ['how', 'what', 'why', 'when', 'where', 'who', 'can you', '?'];

            var lowerTranscript = transcript.toLowerCase();
            var confidentCount = confidentWords.filter(function(w) {
                return lowerTranscript.indexOf(w) !== -1;
            }).length;
            var confusedCount = confusedWords.filter(function(w) {
                return lowerTranscript.indexOf(w) !== -1;
            }).length;
            var questionCount = questionWords.filter(function(w) {
                return lowerTranscript.indexOf(w) !== -1;
            }).length;

            if (confusedCount > 0) {
                analysis.sentiment = 'confused';
            } else if (confidentCount > 0) {
                analysis.sentiment = 'confident';
            } else if (questionCount > 0) {
                analysis.sentiment = 'engaged';
            }

            // Understanding score
            analysis.understandingScore = 0.5;
            if (analysis.sentiment === 'confident') analysis.understandingScore += 0.3;
            if (analysis.sentiment === 'confused') analysis.understandingScore -= 0.2;
            if (analysis.sentiment === 'engaged') analysis.understandingScore += 0.1;
            if (analysis.fillers.length > 3) analysis.understandingScore -= 0.1;
            analysis.understandingScore = Math.max(0, Math.min(1, analysis.understandingScore));

            return analysis;
        },

        // Fallback to server-side processing
        fallbackToServer: function() {
            this.log('Using server-side voice processing');
            // Server endpoints will handle STT/TTS
            this.useServer = true;
        },

        // Server-side TTS
        speakServer: function(text, emotion) {
            var self = this;
            fetch('/api/voice/tts', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    text: text,
                    emotion: emotion || 'neutral'
                })
            })
            .then(function(response) {
                if (!response.ok) throw new Error('TTS failed');
                return response.blob();
            })
            .then(function(blob) {
                var audio = new Audio(URL.createObjectURL(blob));
                audio.play();
                self.isSpeaking = true;
                audio.onended = function() {
                    self.isSpeaking = false;
                    self.updateSpeakerButton(false);
                };
            })
            .catch(function(err) {
                self.log('Server TTS error:', err);
            });
        },

        log: function() {
            if (this.options.debug) {
                console.log.apply(console, ['[VoiceChat]'].concat(Array.prototype.slice.call(arguments)));
            }
        }
    };

    // Export to window
    window.VoiceChat = VoiceChat;

    // Auto-initialize if data attributes present
    document.addEventListener('DOMContentLoaded', function() {
        var chatContainer = document.querySelector('[data-voice-chat]');
        if (chatContainer) {
            var voiceChat = new VoiceChat({
                debug: chatContainer.getAttribute('data-voice-debug') === 'true'
            });
            window.voiceChat = voiceChat;
        }
    });

})(window);

/**
 * Usage Examples:
 *
 * // Basic usage (auto-initializes with data-voice-chat)
 * <div data-voice-chat data-voice-debug="true">
 *     <input id="chat-input" type="text">
 *     <button data-voice-mic>🎤</button>
 *     <button data-voice-speaker>🔊</button>
 * </div>
 *
 * // Manual initialization
 * var voice = new VoiceChat({
 *     inputSelector: '#chat-input',
 *     micButtonSelector: '#mic-btn',
 *     onTranscript: function(text) {
 *         console.log('Transcribed:', text);
 *     },
 *     debug: true
 * });
 *
 * // Start/stop listening
 * voice.startListening();
 * voice.stopListening();
 *
 * // Text-to-speech
 * voice.speak('Hello, student!', { emotion: 'excited' });
 * voice.stopSpeaking();
 *
 * // Analyze speech
 * var analysis = voice.analyzeSpeech('I think I understand now', { duration: 3000 });
 * console.log(analysis.understandingScore);
 */
