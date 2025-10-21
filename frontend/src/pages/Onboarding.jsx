import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import Navbar from "../components/Navbar";
import logo from "../logo.png";
import { BrowserOpenURL } from "../../wailsjs/runtime";
import {
  FormatSystemDetails,
  GetPlainSystemDetails,
  GetSettings,
  NavigateExplorerToLogLocation,
} from "../../wailsjs/go/backend/Backend";
import {
  SaveStoreHighlights,
  SaveToken,
} from "../../wailsjs/go/backend/Settings";
import { toast } from "react-hot-toast";

export default function Onboarding() {
  const [loaded, setLoadState] = useState(false);
  const [onboardingComplete, setOnboardingComplete] = useState(false);
  const [token, setToken] = useState("");
  const [coversUploading, setCoversUploading] = useState(false);
  const [storeHighlights, setStoreHighlights] = useState(false); // default on as users run into this issue more than not but give ample warning
  const [tokenInput, setTokenInput] = useState("");
  const [systemDetails, setSystemDetails] = useState(
    "Fetching system details...",
  );

  useEffect(() => {
    GetSettings().then((settings) => {
      setLoadState(true);
      if (settings.notado_token !== "") {
        setOnboardingComplete(true);
      }
      setToken(settings.notado_token);
      setTokenInput(settings.notado_token);
      setStoreHighlights(settings.upload_store_highlights);
    });
    GetPlainSystemDetails().then((details) => setSystemDetails(details));
  }, [loaded]);

  function saveAllSettings() {
    if (tokenInput === "") {
      toast.error("Please enter your Notado token");
      return;
    }
    SaveToken(tokenInput);
    SaveStoreHighlights(storeHighlights);
    navigate("/selector");
  }

  function checkTokenValid() {
    toast.promise(CheckTokenValidity(tokenInput), {
      loading: "Contacting Notado...",
      success: () => "Your API token is valid!",
      error: (err) => {
        if (err === "401 Unauthorized") {
          return "Notado rejected your token";
        }
        return err;
      },
    });
  }

  const navigate = useNavigate();

  if (onboardingComplete) {
    navigate("/selector");
  }
  return (
    <div className="min-h-screen bg-gray-100 dark:bg-gray-800 flex flex-col">
      <Navbar />
      <div className="flex-grow items-center justify-center pb-24 px-24 grid grid-cols-2 gap-14">
        <div className="space-y-2">
          <img
            className="mx-auto h-36 w-auto logo-animation"
            src={logo}
            alt="The October logo, which is a cartoon octopus reading a book."
          />
          <h2 className="text-center text-3xl font-extrabold text-gray-900 dark:dark:text-gray-300">
            First time setup with Noctober
          </h2>
          <p className="mt-0 text-center text-sm text-gray-600 dark:text-gray-400">
            This should only take a minute of your time
          </p>
        </div>
        <div className="space-y-4">
          <div className="bg-white dark:bg-slate-700 shadow sm:rounded-lg">
            <div className="px-4 py-5 sm:p-6">
              <h3 className="text-lg leading-6 font-medium text-gray-900 dark:text-gray-300">
                Set your Notado access token
              </h3>
              <div className="mt-2 max-w-xl text-sm text-gray-500 dark:text-gray-400">
                <p>
                  You can find your access token at{" "}
                  <button
                    className="text-gray-600 dark:text-gray-400 underline"
                    onClick={() =>
                      BrowserOpenURL("https://notado.app/settings")
                    }
                  >
                    https://notado.app/settings
                  </button>
                </p>
              </div>
              <form
                onSubmit={(e) => e.preventDefault()}
                className="sm:flex flex-col"
              >
                <div className="w-full mt-4 sm:flex sm:items-center">
                  <input
                    onChange={(e) => setTokenInput(e.target.value)}
                    type="text"
                    name="token"
                    id="token"
                    className="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full sm:text-sm border-gray-300 dark:bg-gray-200 focus:bg-white rounded-md"
                    placeholder="Your access token goes here"
                    value={tokenInput}
                  />
                </div>
              </form>
            </div>
          </div>
          <div className="bg-white dark:bg-slate-700 shadow sm:rounded-lg">
            <div className="shadow overflow-hidden sm:rounded-md">
              <div className="px-4 py-5 bg-white dark:bg-slate-700 space-y-6 sm:p-6">
                <fieldset>
                  <legend className="text-base font-medium text-gray-900 dark:text-gray-300">
                    All done?
                  </legend>
                  <div className="space-y-4">
                    <div className="flex items-start">
                      <div className="w-full mt-4 sm:flex flex-row">
                        <button
                          onClick={saveAllSettings}
                          type="submit"
                          className="mt-3 w-full inline-flex items-center justify-center px-4 py-2 border border-transparent shadow-sm font-medium rounded-md text-white bg-green-600 hover:bg-green-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-green-500 sm:mt-0 sm:ml-3 sm:text-sm"
                        >
                          Complete setup
                        </button>
                      </div>
                    </div>
                  </div>
                </fieldset>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
