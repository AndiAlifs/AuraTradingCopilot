import { Component, OnInit, ElementRef, ViewChild, AfterViewChecked } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { TradeService, StrategyCardData, ConversationTurn, ModelInfo } from '../../services/trade.service';
import { StrategyCardComponent } from '../strategy-card/strategy-card.component';
import { MarkdownPipe } from '../../pipes/markdown.pipe';

interface ChatMessage {
  role: 'user' | 'aura';
  text?: string;
  isAnalyzing?: boolean;
  strategy?: StrategyCardData;
}

@Component({
  selector: 'app-chat',
  standalone: true,
  imports: [CommonModule, FormsModule, StrategyCardComponent, MarkdownPipe],
  host: {
    class: 'flex flex-col flex-1 overflow-hidden'
  },
  template: `
    <div class="flex flex-col flex-1 overflow-hidden bg-slate-900 w-full max-w-4xl mx-auto border-x border-slate-800 shadow-2xl relative">

      <!-- ── Model selector bar ─────────────────────────────────────── -->
      <div class="flex items-center justify-between px-4 py-2 border-b border-slate-800 bg-slate-950/60 backdrop-blur-sm">
        <div class="flex items-center gap-2">
          <span class="text-xs font-semibold tracking-widest text-slate-500 uppercase">Engine</span>
          <div class="relative model-select-wrapper">
            <select
              [(ngModel)]="selectedModelId"
              (ngModelChange)="onModelChange($event)"
              id="model-selector"
              class="appearance-none bg-slate-800 border border-slate-700 text-slate-200 text-xs font-mono
                     rounded-full pl-3 pr-8 py-1.5 cursor-pointer
                     focus:outline-none focus:border-auraGreen focus:ring-1 focus:ring-auraGreen
                     transition-all hover:border-slate-500"
              [disabled]="models.length === 0">
              <option *ngFor="let m of models" [value]="m.id">{{ m.name }}</option>
            </select>
            <!-- Chevron icon -->
            <svg class="pointer-events-none absolute right-2 top-1/2 -translate-y-1/2 h-3 w-3 text-slate-400"
                 viewBox="0 0 20 20" fill="currentColor">
              <path fill-rule="evenodd"
                    d="M5.23 7.21a.75.75 0 011.06.02L10 10.94l3.71-3.71a.75.75 0 111.06 1.06l-4.24 4.24a.75.75 0 01-1.06 0L5.21 8.27a.75.75 0 01.02-1.06z"
                    clip-rule="evenodd" />
            </svg>
          </div>
        </div>

        <!-- Provider badge -->
        <div *ngIf="currentModel" class="flex items-center gap-1.5">
          <span class="inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[10px] font-semibold tracking-wide uppercase"
                [class.bg-blue-900]="currentModel.provider === 'google'"
                [class.text-blue-300]="currentModel.provider === 'google'"
                [class.bg-purple-900]="currentModel.provider === 'ollama'"
                [class.text-purple-300]="currentModel.provider === 'ollama'">
            <!-- Google icon -->
            <svg *ngIf="currentModel.provider === 'google'" class="h-2.5 w-2.5" viewBox="0 0 24 24" fill="currentColor">
              <path d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/>
              <path d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
              <path d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"/>
              <path d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"/>
            </svg>
            <!-- Ollama dot icon -->
            <svg *ngIf="currentModel.provider === 'ollama'" class="h-2.5 w-2.5" viewBox="0 0 24 24" fill="currentColor">
              <circle cx="12" cy="12" r="10"/>
            </svg>
            {{ currentModel.provider === 'google' ? 'Google AI' : 'Ollama (local)' }}
          </span>
        </div>
      </div>

      <!-- ── Message list ────────────────────────────────────────────── -->
      <div class="flex-1 overflow-y-auto p-6 space-y-6" #scrollContainer>

        <div class="text-center py-10 opacity-80">
          <img src="/asset/Aura.png" alt="Aura"
               class="w-32 h-32 rounded-full mx-auto mb-4 border-2 border-slate-700 shadow-lg object-cover"
               onerror="this.src='data:image/svg+xml;utf8,<svg xmlns=\\'http://www.w3.org/2000/svg\\' width=\\'128\\' height=\\'128\\'><rect width=\\'128\\' height=\\'128\\' fill=\\'%231e293b\\'/><text x=\\'64\\' y=\\'80\\' font-size=\\'48\\' font-family=\\'sans-serif\\' text-anchor=\\'middle\\' fill=\\'%234ade80\\'>A</text></svg>'">
          <h2 class="text-2xl font-light text-slate-300 tracking-wide">Hi, I'm <span class="font-bold text-auraGreen">Aura</span>.</h2>
          <p class="text-slate-500 mt-2">Your personal trading assistant. Ask me anything about the market.</p>
        </div>

        <div *ngFor="let msg of messages"
             class="flex flex-col animate-fade-in-up"
             [class.items-end]="msg.role === 'user'"
             [class.items-start]="msg.role === 'aura'">

          <div *ngIf="msg.text"
               class="max-w-[80%] rounded-2xl px-5 py-3 shadow-md text-sm md:text-base leading-relaxed"
               [class.bg-blue-600]="msg.role === 'user'"
               [class.text-white]="msg.role === 'user'"
               [class.bg-auraPanel]="msg.role === 'aura'"
               [class.text-slate-200]="msg.role === 'aura'"
               [class.border]="msg.role === 'aura'"
               [class.border-slate-700]="msg.role === 'aura'"
               [class.prose]="msg.role === 'aura'"
               [class.prose-invert]="msg.role === 'aura'">
            <ng-container *ngIf="msg.role === 'user'">{{msg.text}}</ng-container>
            <div *ngIf="msg.role === 'aura'" class="markdown-body" [innerHTML]="msg.text | markdown"></div>
          </div>

          <div *ngIf="msg.isAnalyzing"
               class="flex items-center space-x-3 bg-slate-800/50 rounded-2xl px-5 py-3 border border-slate-700 mt-2">
            <div class="relative flex h-4 w-4">
              <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-auraGreen opacity-75"></span>
              <span class="relative inline-flex rounded-full h-4 w-4 bg-green-500"></span>
            </div>
            <span class="text-sm text-slate-400 italic">Aura is fetching live market data and analyzing...</span>
          </div>

          <app-strategy-card *ngIf="msg.strategy" [data]="msg.strategy"></app-strategy-card>
        </div>
      </div>

      <!-- ── Input bar ───────────────────────────────────────────────── -->
      <div class="p-4 bg-slate-900 border-t border-slate-800">
        <form (ngSubmit)="sendMessage()" class="relative flex items-center">
          <input type="text" [(ngModel)]="inputMessage" name="inputMessage"
                 #chatInput
                 class="w-full bg-slate-800 border border-slate-700 text-white rounded-full pl-5 pr-12 py-3 focus:outline-none focus:border-auraGreen focus:ring-1 focus:ring-auraGreen transition-all shadow-inner"
                 placeholder="Ask Aura anything — or just drop a ticker...">
          <button type="submit"
                  [disabled]="!inputMessage.trim() || isAnalyzing"
                  class="absolute right-2 top-1/2 -translate-y-1/2 p-2 bg-auraGreen text-slate-900 rounded-full hover:bg-green-400 disabled:opacity-50 disabled:cursor-not-allowed transition-colors">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
              <path d="M10.894 2.553a1 1 0 00-1.788 0l-7 14a1 1 0 001.169 1.409l5-1.429A1 1 0 009 15.571V11a1 1 0 112 0v4.571a1 1 0 00.725.962l5 1.428a1 1 0 001.17-1.408l-7-14z" />
            </svg>
          </button>
        </form>
      </div>
    </div>
  `,
  styles: [`
    .animate-fade-in-up {
      animation: fadeInUp 0.4s ease-out forwards;
    }
    @keyframes fadeInUp {
      from { opacity: 0; transform: translateY(10px); }
      to   { opacity: 1; transform: translateY(0); }
    }
    .model-select-wrapper select option {
      background: #1e293b;
      color: #e2e8f0;
    }
    ::ng-deep .markdown-body p:last-child {
      margin-bottom: 0;
    }
    ::ng-deep .markdown-body p:first-child {
      margin-top: 0;
    }
    ::ng-deep .markdown-body strong {
      color: #fff;
    }
  `]
})
export class ChatComponent implements OnInit, AfterViewChecked {
  messages: ChatMessage[] = [];
  inputMessage = '';
  isAnalyzing = false;

  models: ModelInfo[] = [];
  selectedModelId = 'gemini-2.5-pro';

  get currentModel(): ModelInfo | undefined {
    return this.models.find(m => m.id === this.selectedModelId);
  }

  @ViewChild('scrollContainer') private scrollContainer!: ElementRef;
  @ViewChild('chatInput') private chatInput!: ElementRef;

  constructor(private tradeService: TradeService) {}

  ngOnInit() {
    // Load available models from backend
    this.tradeService.getModels().subscribe({
      next: (models) => {
        this.models = models;
        this.selectedModelId = this.tradeService.selectedModel;
      },
      error: () => {
        // Static fallback so UI still works when backend is down
        this.models = [
          { id: 'gemini-2.5-pro',                name: 'Gemini 2.5 Pro',        provider: 'google' },
          { id: 'gemini-3.1-pro-preview',        name: 'Gemini 3.1 Pro',        provider: 'google' },
          { id: 'gemini-3-flash-preview',        name: 'Gemini 3 Flash',        provider: 'google' },
          { id: 'gemini-3.1-flash-lite-preview', name: 'Gemini 3.1 Flash Lite', provider: 'google' },
          { id: 'kimi-k2.6',                     name: 'Kimi K2.6',             provider: 'ollama' },
          { id: 'deepseek-v4-pro',               name: 'DeepSeek V4 Pro',       provider: 'ollama' },
          { id: 'gemma4',                        name: 'Gemma 4',               provider: 'ollama' },
          { id: 'glm-5.1',                       name: 'GLM 5.1',               provider: 'ollama' },
        ];
      }
    });

    this.tradeService.chatInputFocus$.subscribe(ticker => {
      if (ticker) {
        this.inputMessage = `What's the setup for ${ticker}?`;
        setTimeout(() => {
          this.chatInput.nativeElement.focus();
          this.sendMessage();
        }, 0);
        this.tradeService.clearChatTrigger();
      }
    });
  }

  onModelChange(modelId: string) {
    this.tradeService.setModel(modelId);
  }

  ngAfterViewChecked() {
    this.scrollToBottom();
  }

  scrollToBottom(): void {
    try {
      this.scrollContainer.nativeElement.scrollTop = this.scrollContainer.nativeElement.scrollHeight;
    } catch { }
  }

  sendMessage() {
    if (!this.inputMessage.trim() || this.isAnalyzing) return;

    const userText = this.inputMessage.trim();

    // Capture history before adding the new user message
    const history: ConversationTurn[] = this.messages
      .filter(m => m.text && !m.isAnalyzing)
      .map(m => ({ role: m.role, text: m.text! }));

    this.messages.push({ role: 'user', text: userText });
    this.inputMessage = '';
    this.isAnalyzing = true;
    this.messages.push({ role: 'aura', isAnalyzing: true });

    this.tradeService.chat(history, userText).subscribe({
      next: (resp) => {
        this.messages = this.messages.filter(m => !m.isAnalyzing);
        this.isAnalyzing = false;

        if (resp.text) {
          this.messages.push({ role: 'aura', text: resp.text });
        }
        if (resp.strategy) {
          this.messages.push({ role: 'aura', strategy: resp.strategy });
        }
        if (!resp.text && !resp.strategy) {
          this.messages.push({ role: 'aura', text: "I didn't get a clear response. Try again." });
        }
      },
      error: (err) => {
        this.messages = this.messages.filter(m => !m.isAnalyzing);
        this.isAnalyzing = false;
        const msg = err?.status === 401
          ? "You need to sign in to use Aura. Head to /login to get started."
          : "Something went wrong on my end — check that the backend is running and try again.";
        this.messages.push({ role: 'aura', text: msg });
      }
    });
  }
}
