// http server for sif.io
package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", index(index_html))
	http.HandleFunc("/resume", page(resume_html))
	log.Fatal(http.ListenAndServe("0.0.0.0:8000", nil))
}

func index(i string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/" {
			log.Printf("path=%q ip=%q status=404", req.URL.Path, req.RemoteAddr)
			http.NotFound(w, req)
			return
		}
		page(i)(w, req)
	}
}

func page(page string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		log.Printf("path=%q ip=%q status=200", req.URL.Path, req.RemoteAddr)
		fmt.Fprint(w, page)
	}
}

var index_html = `<!DOCTYPE html>
    <html>
    <head>
    	<meta charset="UTF-8">
    	<meta name="viewport" content="width=device-width, initial-scale=1">
    	<title>sifio</title>
    	<style>
    		.flex-container {
    			display: -webkit-flex;
    			display: flex;
    			-webkit-flex-flow: row wrap;
    			flex-flow: row wrap;
    			text-align: left;
    		}
    		.flex-container > * {
    			padding: 15px;
    			-webkit-flex: 1 100%;
    			flex: 1 100%;
    		}
    		.article {
    			text-align: left;
    			background: #3A6EA5;
    			color: white;
    		}
    		header {background: #dd9c37; color: white;}
    		.nav {background: azure;}
    		.nav ul {
    			list-style-type: none;
    			padding: 0;
    		}
    		.nav ul a {
    		}
    		@media all and (min-width: 768px) {
    			.nav {text-align:left;-webkit-flex: 1 auto;flex:1 auto;-webkit-order:1;order:1;}
    			.article {-webkit-flex:5 0px;flex:5 0px;-webkit-order:2;order:2;}
    			footer {-webkit-order:3;order:3;}
    		}
    	</style>
    </head>
    <body>
    	<div class="flex-container">
    		<header><h2>Elijah Buck</h2></header>
    		<nav class="nav"><ul>
    			<li><a href="/resume">Resume</a><br></li>
    			<li><a href="https://github.com/buckelij">GitHub</a></li>
    		</ul></nav>
    		<article class="article">
    			elijah.buck [AT] gmail.com </br>
    			buckelij [AT] sif.io </br>
    			<hr>
    			Software Engineer at GitHub
    		</article>
    	</div>
    </body>
    </html>`

var resume_html = `<!DOCTYPE HTML>
	<html>
	<head>
		<meta charset="UTF-8">
		<title>Résumé</title>
	</head>
	<body>
		Elijah Buck <br>
		elijah.buck [AT] gmail.com <br>
		<br>
		<b>Professional Experience</b>
		<ul>
			<li>Staff Software Engineer, Customer Success Engineering, GitHub, 2021-Present</li>
			<li>Software Engineer, Support Operations, GitHub, 2018-2021</li>
			<li>Enterprise Support Engineer, GitHub, 2014-2018</li>
			<li>Systems Administrator, The Jockey Club Technology Services, 2011-2014</li>
			<li>Systems Administrator, The University of Chicago, 2008-2011</li>
		</ul>
		<b>Education</b>
		<ul>
			<li>Bachelor of Arts in Computer Science with Honors, Grinnell College, Grinnell IA</li>
			<li>Concentration in Neuroscience, Grinnell College, Grinnell IA</li>
		</ul>
		<b>Skills</b>
		<ul>
			<li>Languages: Ruby, JavaScript, Go, Bash</li>
			<li>Systems: AWS, Azure, HAProxy, Kubernetes, Linux, MySQL , Terraform, etc :)</li>
		</ul>
	</body>
	</html>`
