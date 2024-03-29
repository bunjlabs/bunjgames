# Whirligig game file specification.

Zip archive file with structure:
 - content.xml
 - assets/  - images, audio and video folder

content.xml structure:
~~~
<?xml version="1.0" encoding="utf-8"?>
<!DOCTYPE game>
<game>
    <items>  <!-- 13 items -->
        <item>
            <number>1</number>  <!-- integer -->
            <name>1</name>  <!-- string -->
            <description>question</description>  <!-- string -->
            <type>standard</type>  <!-- standard, blitz, superblitz -->
            <questions> <!-- 1 for standard, 3 for blitz and superblitz -->
                <question>
                    <description>question</description>  <!-- string -->
                    <text></text>  <!-- string, optional -->
                    <image></image>  <!-- string, optional -->
                    <audio></audio>  <!-- string, optional -->
                    <video></video>  <!-- string, optional -->
                    <answer>
                        <description>answer</description>  <!-- string -->
                        <text></text>  <!-- string, optional -->
                        <image></image>  <!-- string, optional -->
                        <audio></audio>  <!-- string, optional -->
                        <video></video>  <!-- string, optional -->
                    </answer>
                </question>
                ...  <!-- 1 item for standard question, 3 for blitz and superblitz -->
            </questions>
        </item>
   </items>
   ...
</game>
~~~


# Jeopardy

game packs editor - https://vladimirkhil.com/si/siquester

game packs - https://vladimirkhil.com/si/storage

# Weakest

Text (XML) file with following structure:
~~~
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE game>
<game>
   <questions> <!-- Should contain a lot of questions, recommended amount is 100 - 200 -->
      <question>
         <question>question</question> <!-- string -->
         <answer>answer</answer> <!-- string -->
      </question>
      ...
   </questions>
   <final_questions> <!-- Minimum 10 questions, recommended 20 -->
      <question>
         <question>question</question> <!-- string -->
         <answer>answer</answer> <!-- string -->
      </question>
      ...
   </final_questions>
   <score_multiplier>1</score_multiplier> <!-- integer, determines the score multiplyer -->
</game>
~~~