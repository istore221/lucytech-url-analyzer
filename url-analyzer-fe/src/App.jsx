import React, { useState, useEffect, useRef } from 'react';
import { Globe, Play, Copy, Check, RefreshCw, AlertCircle, Sparkles, Terminal, ChevronsDown, ChevronsUp, XCircle } from 'lucide-react';
import { JsonViewer } from '@textea/json-viewer';

export default function App() {
  const [url, setUrl] = useState('https://www.scrapethissite.com/login/');
  const [loading, setLoading] = useState(false);
  
  // Timeout settings fetched from API
  const [defaultTimeout, setDefaultTimeout] = useState(20);
  const [timeLeft, setTimeLeft] = useState(20);

  const [responseJson, setResponseJson] = useState(null);
  const [error, setError] = useState(null);
  const [copied, setCopied] = useState(false);
  
  // Default to expanded state
  const [isAllExpanded, setIsAllExpanded] = useState(true);

  const timerRef = useRef(null);

  // Fetch timeout configuration on component mount
  useEffect(() => {
    const fetchConfig = async () => {
      try {
        const res = await fetch('http://localhost:8080/api/v1/config');
        if (res.ok) {
          const data = await res.json();
          if (data && typeof data.scrape_timeout_seconds === 'number') {
            setDefaultTimeout(data.scrape_timeout_seconds);
            setTimeLeft(data.scrape_timeout_seconds);
          }
        }
      } catch (err) {
        console.warn('Could not fetch config from backend, using default 20s fallback:', err);
      }
    };

    fetchConfig();
  }, []);

  const handleScrape = async (e) => {
    e.preventDefault();
    
    const trimmedUrl = url.trim();
    if (!trimmedUrl) {
      setError({
        code: 'VALIDATION_ERROR',
        message: 'Please enter a valid URL.',
      });
      return;
    }

    setLoading(true);
    setError(null);
    setResponseJson(null);
    
    // Reset timer and state
    const startDuration = defaultTimeout;
    setTimeLeft(startDuration);
    setIsAllExpanded(true);

    if (timerRef.current) clearInterval(timerRef.current);
    
    timerRef.current = setInterval(() => {
      setTimeLeft((prev) => {
        if (prev <= 1) {
          clearInterval(timerRef.current);
          return 0;
        }
        return prev - 1;
      });
    }, 1000);

    try {
      const payload = {
        urls: [
          {
            url: trimmedUrl
          }
        ]
      };

      const res = await fetch('http://localhost:8080/api/v1/scrape', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(payload),
      });

      const data = await res.json();
      setResponseJson(data);

      // 1. Check for top-level batch/API error (e.g. status 400 or top-level "error" object)
      if (!res.ok || data.error) {
        setError(data.error || { code: `HTTP_${res.status}`, message: res.statusText });
        return;
      }

      // 2. Check for item-level URL processing errors inside data array (status 200 with error payload)
      if (Array.isArray(data.data) && data.data.length > 0) {
        const itemWithError = data.data.find((item) => item.error);
        if (itemWithError) {
          setError({
            url: itemWithError.url,
            ...itemWithError.error
          });
        }
      }
    } catch (err) {
      setError({
        code: 'CONNECTION_ERROR',
        message: err.message || 'Failed to connect to the API server.',
      });
    } finally {
      if (timerRef.current) {
        clearInterval(timerRef.current);
      }
      setLoading(false);
    }
  };

  const handleCopy = () => {
    if (responseJson) {
      navigator.clipboard.writeText(JSON.stringify(responseJson, null, 2));
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  };

  useEffect(() => {
    return () => {
      if (timerRef.current) clearInterval(timerRef.current);
    };
  }, []);

  const progressPercentage = defaultTimeout > 0 
    ? ((defaultTimeout - timeLeft) / defaultTimeout) * 100 
    : 100;

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 font-sans flex flex-col items-center justify-center py-12 px-4 sm:px-6 lg:px-8">
      {/* Background Glows */}
      <div className="fixed top-1/4 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[600px] h-[300px] bg-indigo-600/15 blur-[120px] pointer-events-none rounded-full" />
      <div className="fixed bottom-10 right-10 w-80 h-80 bg-cyan-600/10 blur-[100px] pointer-events-none rounded-full" />

      {/* Main Container */}
      <div className="w-full max-w-3xl relative z-10 space-y-8">
        
        {/* Header */}
        <div className="text-center space-y-3">
          <div className="inline-flex items-center gap-2 px-3.5 py-1.5 rounded-full bg-slate-900 border border-slate-800 text-xs text-indigo-400 font-medium">
            <Sparkles className="w-3.5 h-3.5" /> lucytech-url-analyzer-fe
          </div>
          <h1 className="text-4xl sm:text-5xl font-extrabold tracking-tight bg-gradient-to-r from-white via-slate-200 to-slate-400 bg-clip-text text-transparent">
            Lucytech Assignment
          </h1>
          <p className="text-slate-400 text-sm sm:text-base max-w-lg mx-auto">
            Enter a URL below to process.
          </p>
        </div>

        {/* Input Card */}
        <div className="bg-slate-900/90 border border-slate-800 backdrop-blur-xl rounded-2xl p-6 sm:p-8 shadow-2xl space-y-6">
          <form onSubmit={handleScrape} className="space-y-4">
            <div className="flex flex-col space-y-2">
              <label className="text-xs font-semibold uppercase tracking-wider text-slate-400 flex justify-between items-center">
                <span>Target Web Page URL</span>
              </label>

              <div className="relative flex items-center">
                <Globe className="absolute left-4 w-5 h-5 text-slate-500 pointer-events-none" />
                <input
                  type="url"
                  required
                  placeholder="https://www.scrapethissite.com/login/"
                  value={url}
                  onChange={(e) => setUrl(e.target.value)}
                  disabled={loading}
                  className="w-full bg-slate-950 border border-slate-800 focus:border-indigo-500 focus:ring-2 focus:ring-indigo-500/20 text-slate-100 placeholder-slate-600 rounded-xl pl-12 pr-36 py-4 text-sm sm:text-base transition-all outline-none"
                />
                <button
                  type="submit"
                  disabled={loading}
                  className="absolute right-2 top-2 bottom-2 px-5 rounded-lg bg-indigo-600 hover:bg-indigo-500 active:bg-indigo-700 text-white font-medium text-sm transition-all flex items-center justify-center gap-2 shadow-md shadow-indigo-600/30 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {loading ? (
                    <>
                      <RefreshCw className="w-4 h-4 animate-spin" />
                      <span>Processing</span>
                    </>
                  ) : (
                    <>
                      <Play className="w-4 h-4 fill-current" />
                      <span>Process</span>
                    </>
                  )}
                </button>
              </div>
            </div>
          </form>

          {/* Loading Progress & Countdown Bar */}
          {loading && (
            <div className="pt-4 border-t border-slate-800 space-y-3">
              <div className="flex items-center justify-between text-xs font-mono">
                <span className="text-slate-400 flex items-center gap-2">
                  <RefreshCw className="w-3.5 h-3.5 animate-spin text-indigo-400" />
                  Fetching response from server...
                </span>
                <span className="text-indigo-400 font-bold bg-indigo-500/10 px-2.5 py-1 rounded-md border border-indigo-500/20">
                  {timeLeft}s remaining
                </span>
              </div>

              <div className="w-full bg-slate-950 h-2.5 rounded-full overflow-hidden border border-slate-800">
                <div
                  className="bg-gradient-to-r from-indigo-500 to-cyan-400 h-full transition-all duration-1000 ease-linear rounded-full"
                  style={{ width: `${progressPercentage}%` }}
                />
              </div>
            </div>
          )}

          {/* Structured Error Alert Card */}
          {error && (
            <div className="p-4 rounded-xl bg-red-950/40 border border-red-500/30 text-red-300 space-y-2">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2 font-semibold text-red-400 text-sm">
                  <AlertCircle className="w-4 h-4 text-red-400 shrink-0" />
                  <span>Processing Error</span>
                </div>
                {error.code && (
                  <span className="text-[10px] font-mono uppercase bg-red-500/20 text-red-300 px-2 py-0.5 rounded border border-red-500/30">
                    {error.code}
                  </span>
                )}
              </div>

              {error.message && (
                <p className="text-xs text-slate-300 pl-6 leading-relaxed">
                  {error.message}
                </p>
              )}

              {error.url && (
                <div className="pl-6 text-[11px] font-mono text-slate-400 truncate">
                  <span className="text-slate-500">Target URL: </span>{error.url}
                </div>
              )}

              {Array.isArray(error.details) && error.details.length > 0 && (
                <div className="pl-6 pt-1">
                  <ul className="list-disc list-inside text-xs text-slate-400 space-y-1 bg-slate-950/50 p-2.5 rounded-lg border border-red-500/10">
                    {error.details.map((detail, idx) => (
                      <li key={idx}>{detail}</li>
                    ))}
                  </ul>
                </div>
              )}
            </div>
          )}
        </div>

        {/* Interactive JSON Viewer */}
        {responseJson && (
          <div className="bg-slate-900/90 border border-slate-800 backdrop-blur-xl rounded-2xl overflow-hidden shadow-2xl">
            <div className="flex items-center justify-between px-6 py-4 bg-slate-950/80 border-b border-slate-800">
              <div className="flex items-center gap-2 text-slate-300 font-mono text-xs">
                <Terminal className="w-4 h-4 text-emerald-400" />
                <span>Response Tree</span>
                <span className={`px-2 py-0.5 rounded text-[10px] font-semibold border ${
                  error 
                    ? 'bg-amber-500/10 text-amber-400 border-amber-500/20' 
                    : 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20'
                }`}>
                  {responseJson.status ? `${responseJson.status}` : '200 OK'}
                </span>
              </div>

              {/* Header Actions */}
              <div className="flex items-center gap-2">
                <button
                  onClick={() => setIsAllExpanded(!isAllExpanded)}
                  className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 hover:text-white text-xs font-medium transition-all border border-slate-700/50"
                  title={isAllExpanded ? "Collapse deep nodes" : "Expand all nested nodes"}
                >
                  {isAllExpanded ? (
                    <>
                      <ChevronsUp className="w-3.5 h-3.5 text-indigo-400" /> Collapse All
                    </>
                  ) : (
                    <>
                      <ChevronsDown className="w-3.5 h-3.5 text-indigo-400" /> Expand All
                    </>
                  )}
                </button>

                <button
                  onClick={handleCopy}
                  className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 hover:text-white text-xs font-medium transition-all border border-slate-700/50"
                >
                  {copied ? (
                    <>
                      <Check className="w-3.5 h-3.5 text-emerald-400" /> Copied
                    </>
                  ) : (
                    <>
                      <Copy className="w-3.5 h-3.5" /> Copy JSON
                    </>
                  )}
                </button>
              </div>
            </div>

            <div className="p-6 bg-slate-950 overflow-x-auto max-h-[550px]">
              <JsonViewer
                key={isAllExpanded ? 'expanded' : 'collapsed'}
                value={responseJson}
                theme="dark"
                defaultInspectDepth={isAllExpanded ? Infinity : 0}
                rootName={false}
                style={{
                  backgroundColor: 'transparent',
                  fontFamily: 'ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace',
                  fontSize: '0.85rem',
                }}
              />
            </div>
          </div>
        )}

      </div>
    </div>
  );
}