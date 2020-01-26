 
<html>
  <head>
    <title>Guestbook</title>
  </head>
  <body>
    <div style="width: 50%; margin-left: 20px">
      <h2>Guestbook</h2>
      <form action = "<?php $_PHP_SELF ?>" method = "POST">
         Name: <input type = "text" name = "name" />
         Text: <input type = "text" name = "text" />
         <input type = "submit" />
      </form>
    <div>
        <?php
            $api_url = getenv("API_URL") . "/messages";
            echo "Message history:";
            $data = json_decode(file_get_contents($api_url), true);
            $sep = '  ,  ';
            echo '<br>';
            foreach($data as $item) {
                echo $item['date'], $sep, $item['name'], $sep, $item['text'];
                echo '<br>';
            }
        ?>
      </div>
    </div>
    </div>
  </body>
</html>


<?php

if( $_POST["name"] || $_POST["text"] ) {

    //POST DATA
    $name = $_POST['name'];
    $text = $_POST['text'];

    $jsonData = array(
        'Name' =>  $name,
        'Text' =>  $text
    );

    $jsonDataEncoded = json_encode($jsonData);

    //API Url
    $api_url = getenv("API_URL") . "/messages";
    $ch = curl_init();


    curl_setopt($ch, CURLOPT_VERBOSE, true);
    $verbose = fopen('php://temp', 'w+');
    curl_setopt($ch, CURLOPT_STDERR, $verbose);


    curl_setopt($ch, CURLOPT_URL, $api_url);
   // curl_setopt($ch, CURLOPT_POST, 1);
    curl_setopt($ch, CURLOPT_POSTFIELDS, $jsonDataEncoded);
    curl_setopt($ch, CURLOPT_HTTPHEADER, array('Content-Type: application/json')); 
    curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
    $result = curl_exec($ch);
    // echo $api_url;
    // if(!$result){die("Connection Failure");}
    // curl_close($ch);
   
    // return $result;

    $info = curl_getinfo($ch);
    print_r( $info );

    $result = curl_exec($ch);
    if ($result === FALSE) {
        printf("cUrl error (#%d): %s<br>\n", curl_errno($ch),
            htmlspecialchars(curl_error($ch)));
    }

    rewind($verbose);
    $verboseLog = stream_get_contents($verbose);

    echo "Verbose information:\n<pre>", htmlspecialchars($verboseLog), "</pre>\n";


    exit();
 }

?>