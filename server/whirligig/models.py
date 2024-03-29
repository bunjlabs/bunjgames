import datetime
import time

from django.utils import timezone
import random
from xml.etree import ElementTree

from django.db import models, transaction

from common.utils import generate_token, BadFormatException, BadStateException, NothingToDoException


class Game(models.Model):
    STATE_START = 'start'
    STATE_INTRO = 'intro'
    STATE_QUESTIONS = 'questions'
    STATE_QUESTION_WHIRLIGIG = 'question_whirligig'
    STATE_QUESTION_START = 'question_start'
    STATE_QUESTION_DISCUSSION = 'question_discussion'
    STATE_ANSWER = 'answer'
    STATE_RIGHT_ANSWER = 'right_answer'
    STATE_QUESTION_END = 'question_end'
    STATE_END = 'end'

    STATES = (
        STATE_START,
        STATE_INTRO,
        STATE_QUESTIONS,
        STATE_QUESTION_WHIRLIGIG,
        STATE_QUESTION_START,
        STATE_QUESTION_DISCUSSION,
        STATE_ANSWER,
        STATE_RIGHT_ANSWER,
        STATE_QUESTION_END,
        STATE_END,
    )

    CHOICES_STATE = ((o, o) for o in STATES)

    MAX_SCORE = 6

    token = models.CharField(max_length=25, null=True, blank=True, db_index=True)
    created = models.DateTimeField(auto_now_add=True)
    expired = models.DateTimeField()
    connoisseurs_score = models.IntegerField(default=0)
    viewers_score = models.IntegerField(default=0)
    cur_random_item = models.IntegerField(default=None, null=True)
    cur_item = models.IntegerField(default=None, null=True)
    cur_question = models.IntegerField(default=None, null=True)
    state = models.CharField(max_length=25, choices=CHOICES_STATE, default=STATE_START)
    timer_paused = models.BooleanField(default=True)
    timer_paused_time = models.BigIntegerField(default=0)
    timer_time = models.BigIntegerField(default=0)

    def get_curr_item(self):
        return self.items.get(number=self.cur_item)

    def generate_token(self):
        self.token = generate_token(self.pk)
        self.save(update_fields=['token'])

    def set_timer(self, t, save=False):
        self.timer_paused = t == 0
        self.timer_paused_time = 0
        self.timer_time = int(round((time.time() + t + 2) * 1000)) if t > 0 else 0
        if save:
            self.save(update_fields=['timer_paused', 'timer_paused_time', 'timer_time'])

    def clear_timer(self, save=False):
        self.set_timer(0, save)

    @staticmethod
    @transaction.atomic(savepoint=False)
    def new():
        game = Game.objects.create(
            expired=timezone.now() + datetime.timedelta(hours=12)
        )
        game.generate_token()
        return game

    @transaction.atomic(savepoint=False)
    def change_score(self, connoisseurs_score, viewers_score):
        if 0 <= connoisseurs_score <= self.MAX_SCORE:
            self.connoisseurs_score = connoisseurs_score
        if 0 <= viewers_score <= self.MAX_SCORE:
            self.viewers_score = viewers_score
        self.save(update_fields=['connoisseurs_score', 'viewers_score'])

    @transaction.atomic(savepoint=False)
    def change_timer(self, paused):
        if self.state != self.STATE_QUESTION_DISCUSSION:
            raise NothingToDoException()
        if paused and not self.timer_paused:
            self.timer_paused_time = int(round(time.time() * 1000))
        elif not paused and self.timer_paused:
            self.timer_time += int(round(time.time() * 1000)) - self.timer_paused_time
            self.timer_paused_time = 0
        self.timer_paused = paused
        self.save(update_fields=['timer_time', 'timer_paused', 'timer_paused_time'])

    @transaction.atomic(savepoint=False)
    def answer_correct(self, is_correct):
        if self.state != self.STATE_RIGHT_ANSWER:
            raise NothingToDoException()
        self.viewers_score += 0 if is_correct else 1

        item = self.items.get(number=self.cur_item)
        question = item.questions.get(number=self.cur_question)
        question.is_processed = True
        question.save()

        if not is_correct or self.cur_question == item.questions.count() - 1:
            self.connoisseurs_score += 1 if is_correct else 0
            item.is_processed = True
            item.save()
            self.cur_random_item = None
            self.cur_item = None
            self.cur_question = None
            if self.connoisseurs_score == self.MAX_SCORE or self.viewers_score == self.MAX_SCORE \
                    or not self.items.filter(is_processed=False).exists():
                self.expired = timezone.now() + datetime.timedelta(minutes=10)
                self.state = self.STATE_END
            else:
                self.state = self.STATE_QUESTION_END
        else:
            self.cur_question += 1
            self.state = self.STATE_QUESTION_START

        self.save()

    @transaction.atomic(savepoint=False)
    def extra_time(self):
        if self.state != self.STATE_ANSWER:
            raise NothingToDoException()
        self.set_timer(self.get_curr_item().get_time())
        self.state = self.STATE_QUESTION_DISCUSSION
        self.save()

    @transaction.atomic(savepoint=False)
    def parse(self, filename, transform_dict=None):
        if transform_dict is None:
            transform_dict = {}

        tree = ElementTree.parse(filename)

        game_xml = tree.getroot()
        items_xml = game_xml.find('items')

        def format_audio_url(url: str):
            return '/' + transform_dict[url[1:]] if url and url.startswith('/') and url[1:] in transform_dict else url

        for item_number, item_xml in enumerate(items_xml.findall('item')):
            if item_number >= 13:
                raise BadFormatException('Too many items')
            item = GameItem.objects.create(
                number=item_number,
                name=item_xml.find('name').text,
                description=item_xml.find('description').text if item_xml.find('description') is not None else '',
                game=self,
                type=item_xml.find('type').text,
            )
            for question_number, question_xml in enumerate(item_xml.find('questions').findall('question')):
                if question_number >= 3:
                    raise BadFormatException('Too many questions')
                answer_xml = question_xml.find('answer')
                author_xml = question_xml.find('author')
                question = Question.objects.create(
                    number=question_number,
                    item=item,
                    description=question_xml.find('description').text,
                    text=question_xml.find('text').text,
                    image=question_xml.find('image').text,
                    audio=format_audio_url(question_xml.find('audio').text),
                    video=question_xml.find('video').text,

                    answer_description=answer_xml.find('description').text,
                    answer_text=answer_xml.find('text').text,
                    answer_image=answer_xml.find('image').text,
                    answer_audio=format_audio_url(answer_xml.find('audio').text),
                    answer_video=answer_xml.find('video').text,

                    author_name=author_xml.find('name').text if author_xml is not None and author_xml.find('name') is not None else '',
                    author_city=author_xml.find('city').text if author_xml is not None and author_xml.find('city') is not None else '',
                )

    def print(self):
        for item in self.items.iterator():
            print('Item №{}: name={}, type={}'.format(item.number, item.name, item.type))
            for question in item.questions.iterator():
                print('\tQuestion №{}:'.format(question.number))
                print('\t\tdescription: {}'.format(question.description[:50]))
                print('\t\ttext: {}'.format(question.text[:50] if question.text else None))
                print('\t\timage: {}'.format(question.image))
                print('\t\taudio: {}'.format(question.audio))
                print('\t\tvideo: {}'.format(question.video))
                print()
                print('\tAnswer:')
                print('\t\tdescription: {}'.format(question.answer_description[:50]))
                print('\t\ttext: {}'.format(question.answer_text[:50] if question.answer_text else None))
                print('\t\timage: {}'.format(question.answer_image))
                print('\t\taudio: {}'.format(question.answer_audio))
                print('\t\tvideo: {}'.format(question.answer_video))

    def randomise_next_item(self):
        def normalize(a, norm):
            normalized = a % norm
            return norm + normalized if normalized < 0 else normalized

        items = self.items.all()
        random_next_item_number = random.choice(items.values_list('number', flat=True))
        if items.filter(is_processed=False).count() == 0:
            raise BadStateException('No items left')

        cur_number = random_next_item_number
        while items[cur_number].is_processed:
            cur_number = normalize(cur_number + 1, items.count())
        return random_next_item_number, cur_number

    @transaction.atomic(savepoint=False)
    def next_state(self, from_state=None):
        if from_state is not None and self.state != from_state:
            raise NothingToDoException()
        if self.state == self.STATE_START:
            self.state = self.STATE_INTRO
        elif self.state == self.STATE_INTRO:
            self.state = self.STATE_QUESTIONS
        elif self.state == self.STATE_QUESTIONS:
            self.cur_random_item, self.cur_item = self.randomise_next_item()
            self.cur_question = 0
            self.state = self.STATE_QUESTION_WHIRLIGIG
        elif self.state == self.STATE_QUESTION_WHIRLIGIG:
            self.state = self.STATE_QUESTION_START
        elif self.state == self.STATE_QUESTION_START:
            self.set_timer(self.get_curr_item().get_time())
            self.state = self.STATE_QUESTION_DISCUSSION
        elif self.state == self.STATE_QUESTION_DISCUSSION:
            self.clear_timer()
            self.state = self.STATE_ANSWER
        elif self.state == self.STATE_ANSWER:
            self.clear_timer()
            self.state = self.STATE_RIGHT_ANSWER
        elif self.state == self.STATE_RIGHT_ANSWER:
            raise NothingToDoException()
        elif self.state == self.STATE_QUESTION_END:
            self.cur_random_item, self.cur_item = self.randomise_next_item()
            self.cur_question = 0
            self.state = self.STATE_QUESTION_WHIRLIGIG
        elif self.state == self.STATE_END:
            raise NothingToDoException()
        else:
            raise BadStateException('Bad state')
        self.save(update_fields=[
            'state', 'cur_random_item', 'cur_item', 'cur_question', 'timer_paused', 'timer_time', 'timer_paused_time'
        ])

    class Meta:
        indexes = [
            models.Index(fields=['token']),
        ]


class GameItem(models.Model):
    TYPE_STANDARD = 'standard'
    TYPE_BLITZ = 'blitz'
    TYPE_SUPERBLITZ = 'superblitz'

    CHOICES_TYPE = (
        (TYPE_STANDARD, TYPE_STANDARD),
        (TYPE_BLITZ, TYPE_BLITZ),
        (TYPE_SUPERBLITZ, TYPE_SUPERBLITZ),
    )

    number = models.IntegerField()
    name = models.CharField(max_length=255)
    description = models.CharField(max_length=255, blank=True)
    game = models.ForeignKey(Game, on_delete=models.CASCADE, related_name='items')
    type = models.CharField(max_length=25, choices=CHOICES_TYPE)
    is_processed = models.BooleanField(default=False)

    class Meta:
        ordering = ['number']

    def get_time(self):
        return 60 if self.type == self.TYPE_STANDARD else 20


class Question(models.Model):
    number = models.IntegerField()
    item = models.ForeignKey(GameItem, on_delete=models.CASCADE, related_name='questions')
    is_processed = models.BooleanField(default=False)

    description = models.TextField()
    text = models.TextField(null=True)
    image = models.CharField(max_length=255, null=True)
    audio = models.CharField(max_length=255, null=True)
    video = models.CharField(max_length=255, null=True)

    answer_description = models.TextField()
    answer_text = models.TextField(null=True)
    answer_image = models.CharField(max_length=255, null=True)
    answer_audio = models.CharField(max_length=255, null=True)
    answer_video = models.CharField(max_length=255, null=True)

    author_name = models.TextField(null=True)
    author_city = models.TextField(null=True)

    class Meta:
        ordering = ['number']
