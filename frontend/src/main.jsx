import React from 'react';
import { createRoot } from 'react-dom/client'
import { HashRouter, Route, Routes } from "react-router-dom"
import { Toaster } from 'react-hot-toast'

import DeviceSelector from './pages/DeviceSelector';
import Overview from './pages/Overview';
import Settings from './pages/Settings';
import Onboarding from './pages/Onboarding'

import './style.css';

const routes = (
  <React.StrictMode>
    <HashRouter>
      <Routes>
        <Route path="/" element={<Onboarding />} />
        <Route path="/selector" element={<DeviceSelector />} />
        <Route path="/overview" element={<Overview />} />
        <Route path="/settings" element={<Settings />} />
      </Routes>
    </HashRouter>
    <Toaster />
  </React.StrictMode>
)

const container = document.getElementById("root");
const root = createRoot(container);
root.render(routes)
