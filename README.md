# 🤖 Aura — The Swing-Trade Co-Pilot

**Aura** is an agentic AI web application designed for retail stock investors focusing on short-term swing trades and **BSJP (Beli Sore Jual Pagi)** strategies on the Indonesian Stock Exchange (IDX).

## 🌟 Executive Summary
Aura utilizes a Retrieval-Augmented Generation (RAG) agentic workflow. She pulls live market data via RapidAPI, analyzes it using Google's **Gemini 3.1 Pro**, and outputs strict JSON to generate functional UI components for stock market strategies.

## 👩‍💼 Meet Aura
Aura is your Lead System Architect & Quantitative Analyst. 
- **Persona:** Hyper-logical, risk-averse, and highly structured. She speaks in technical certainties and probabilities.
- **Visual Identity:** A professional financial analyst with a dark corporate blazer over a modern slim-fit batik shirt, fitting perfectly with the app's Dark mode (Slate-900/800) with Neon Green and Red accents.

## 🏗️ Technical Architecture
Aura uses a Vercel-optimized Monorepo to maintain security and performance.
- **Frontend:** Angular 18
- **Backend:** Go (Vercel Serverless Functions)
- **AI Model:** Gemini 3.1 Pro
- **Data Provider:** Yahoo Finance (via RapidAPI)

## 🚀 Getting Started

### Prerequisites
- Node.js (v18 or higher)
- Go (v1.21 or higher)
- Angular CLI
- Vercel CLI (recommended for local development)

### Environment Variables
To run this project, the following keys must be configured in your environment (`.env` file or Vercel Environment Variables):

| Variable | Description |
| :--- | :--- |
| `GEMINI_API_KEY` | Your Google AI Studio API Key (for Gemini 3.1 Pro). |
| `RAPIDAPI_KEY` | Your RapidAPI Key for Yahoo Finance data. |
| `RAPIDAPI_HOST` | `yahoo-finance15.p.rapidapi.com` |

### Installation
1. Clone the repository.
2. Install frontend dependencies:
   ```bash
   npm install
   ```
3. Install backend Go modules:
   ```bash
   cd api
   go mod tidy
   ```

### Running Locally (Development)
Since the project relies on Angular for the frontend and Go for backend APIs routing, you can run the app locally using Vercel CLI to simulate the production environment:

```bash
vercel dev
```
Alternatively, run the Angular dev server (note: `/api` endpoints will use the `proxy.conf.json` configuration to forward requests to the Go server if running separately):
```bash
ng serve
```

## 📦 Build & Deployment
Run `ng build` to build the Angular project.
The project is configured to be deployed on **Vercel** out-of-the-box using the provided `vercel.json` routing configuration.

```bash
vercel --prod
```
