import { Component, OnInit, ElementRef, ViewChild, AfterViewChecked } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { TradeService, StrategyCardData } from '../../services/trade.service';
import { StrategyCardComponent } from '../strategy-card/strategy-card.component';

interface ChatMessage {
  role: 'user' | 'aura';
  text?: string;
  isAnalyzing?: boolean;
  strategy?: StrategyCardData;
}

@Component({
  selector: 'app-chat',
  standalone: true,
  imports: [CommonModule, FormsModule, StrategyCardComponent],
  template: `
    <div class="flex flex-col h-full bg-slate-900 w-full max-w-4xl mx-auto border-x border-slate-800 shadow-2xl relative">
      <div class="flex-1 overflow-y-auto p-6 space-y-6" #scrollContainer>
        
        <div class="text-center py-10 opacity-80">
          <img src="/asset/Aura.png" alt="Aura" class="w-32 h-32 rounded-full mx-auto mb-4 border-2 border-slate-700 shadow-lg object-cover" onerror="this.src='data:image/svg+xml;utf8,<svg xmlns=\\\'http://www.w3.org/2000/svg\\\' width=\\\'128\\\' height=\\\'128\\\'><rect width=\\\'128\\\' height=\\\'128\\\' fill=\\\'#1e293b\\\'/><text x=\\\'64\\\' y=\\\'80\\\' font-size=\\\'48\\\' font-family=\\\'sans-serif\\\' text-anchor=\\\'middle\\\' fill=\\\'#4ade80\\\'>A</text></svg>'">
          <h2 class="text-2xl font-light text-slate-300 tracking-wide">Hi, I'm <span class="font-bold text-auraGreen">Aura</span>.</h2>
          <p class="text-slate-500 mt-2">Your Quantitative Analyst for IDX Swing Trading.</p>
        </div>

        <div *ngFor="let msg of messages" class="flex flex-col animate-fade-in-up" [class.items-end]="msg.role === 'user'" [class.items-start]="msg.role === 'aura'">
          <div *ngIf="msg.text" 
               class="max-w-[80%] rounded-2xl px-5 py-3 shadow-md text-sm md:text-base leading-relaxed"
               [class.bg-blue-600]="msg.role === 'user'"
               [class.text-white]="msg.role === 'user'"
               [class.bg-auraPanel]="msg.role === 'aura'"
               [class.text-slate-200]="msg.role === 'aura'"
               [class.border]="msg.role === 'aura'"
               [class.border-slate-700]="msg.role === 'aura'">
            {{msg.text}}
          </div>

          <div *ngIf="msg.isAnalyzing" class="flex items-center space-x-3 bg-slate-800/50 rounded-2xl px-5 py-3 border border-slate-700 mt-2">
            <div class="relative flex h-4 w-4">
              <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-auraGreen opacity-75"></span>
              <span class="relative inline-flex rounded-full h-4 w-4 bg-green-500"></span>
            </div>
            <span class="text-sm text-slate-400 italic">Aura is fetching live market data and analyzing...</span>
          </div>

          <app-strategy-card *ngIf="msg.strategy" [data]="msg.strategy"></app-strategy-card>
        </div>
      </div>

      <div class="p-4 bg-slate-900 border-t border-slate-800">
        <form (ngSubmit)="sendMessage()" class="relative flex items-center">
          <input type="text" [(ngModel)]="inputMessage" name="inputMessage"
                 #chatInput
                 class="w-full bg-slate-800 border border-slate-700 text-white rounded-full pl-5 pr-12 py-3 focus:outline-none focus:border-auraGreen focus:ring-1 focus:ring-auraGreen transition-all shadow-inner" 
                 placeholder="Give me the setup for [Ticker]...">
          <button type="submit" [disabled]="!inputMessage.trim() || isAnalyzing"
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
      to { opacity: 1; transform: translateY(0); }
    }
  `]
})
export class ChatComponent implements OnInit, AfterViewChecked {
  messages: ChatMessage[] = [];
  inputMessage = '';
  isAnalyzing = false;
  
  @ViewChild('scrollContainer') private scrollContainer!: ElementRef;
  @ViewChild('chatInput') private chatInput!: ElementRef;

  constructor(private tradeService: TradeService) {}

  ngOnInit() {
    this.tradeService.chatInputFocus$.subscribe(ticker => {
      this.inputMessage = `Give me the setup for ${ticker}`;
      this.chatInput.nativeElement.focus();
    });
  }

  ngAfterViewChecked() {
    this.scrollToBottom();
  }

  scrollToBottom(): void {
    try {
      this.scrollContainer.nativeElement.scrollTop = this.scrollContainer.nativeElement.scrollHeight;
    } catch(err) { }
  }

  sendMessage() {
    if (!this.inputMessage.trim() || this.isAnalyzing) return;

    const userText = this.inputMessage.trim();
    this.messages.push({ role: 'user', text: userText });
    this.inputMessage = '';
    
    // Extract ticker heuristically
    const words = userText.split(' ');
    let ticker = words[words.length - 1].toUpperCase();
    if (userText.toLowerCase().includes('for')) {
       ticker = userText.split(/for/i)[1].trim().toUpperCase();
    }
    ticker = ticker.replace(/[^A-Z0-9.]/g, '');

    if (!ticker) {
        this.messages.push({ role: 'aura', text: "I couldn't identify a valid ticker in your message." });
        return;
    }

    this.isAnalyzing = true;
    const loadingMsg: ChatMessage = { role: 'aura', isAnalyzing: true };
    this.messages.push(loadingMsg);

    this.tradeService.analyzeTicker(ticker).subscribe({
      next: (data) => {
        this.messages = this.messages.filter(m => !m.isAnalyzing);
        this.isAnalyzing = false;
        
        this.messages.push({ role: 'aura', text: `Here is the quantitative setup for ${data.ticker}:` });
        this.messages.push({ role: 'aura', strategy: data });
      },
      error: (err) => {
        this.messages = this.messages.filter(m => !m.isAnalyzing);
        this.isAnalyzing = false;
        this.messages.push({ role: 'aura', text: `Error analyzing ${ticker}: ${err.message || 'Server error'}` });
      }
    });
  }
}
